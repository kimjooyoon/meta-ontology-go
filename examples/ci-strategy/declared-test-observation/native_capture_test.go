package declaredtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"
)

func nativeCapture(t *testing.T, name string) Capture {
	t.Helper()
	dir := t.TempDir()
	fixture := "package subject\nimport \"testing\"\nfunc TestPass(t *testing.T) {}\n" +
		"func TestDeliberateFail(t *testing.T) { t.Fatal(\"intentional control\") }\n" +
		"func TestSkip(t *testing.T) { t.Skip(\"declared skip\") }\n" +
		"func TestNested(t *testing.T) { t.Run(\"leaf\", func(t *testing.T) {}) }\n"
	module := "module declared-test-subject\n\ngo 1.27.3\n"
	for path, text := range map[string]string{"go.mod": module, "subject_test.go": fixture} {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"),
		"test", "-p", "1", "-json", "-count=1", "-run", "^"+name+"$", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if cmd.ProcessState == nil || ctx.Err() != nil {
		t.Fatal("native process did not finish", err, ctx.Err(), stderr.String())
	}
	exit := cmd.ProcessState.ExitCode()
	revision := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(module+fixture)))
	return Capture{TrialID: "native/" + name, SourceRevision: revision,
		Arguments: slices.Clone(cmd.Args), ExitCode: &exit, Completed: true, StdoutComplete: true, Stdout: stdout.Bytes()}
}

func nativePlan(t *testing.T, capture Capture, test, outcome string, exit int) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"trial_id": capture.TrialID, "source_revision": capture.SourceRevision,
		"process_exit": exit, "tests": []map[string]string{{"test_id": test, "outcome": outcome}}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestRealNamedGoEventsUseSourceAssessment(t *testing.T) {
	for _, row := range []struct {
		name, test, outcome, state string
		exit, observed, matched    int
		failure, accepted          bool
	}{
		{"TestPass", "TestPass", "pass", "MATCHED_EXPECTATION", 0, 1, 1, false, true},
		{"TestDeliberateFail", "TestDeliberateFail", "fail", "MATCHED_EXPECTATION", 1, 1, 1, true, true},
		{"TestAbsent", "TestPass", "pass", "NO_TESTS_OBSERVED", 0, 0, 0, false, false},
		{"TestSkip", "TestSkip", "skip", "MATCHED_EXPECTATION", 0, 1, 1, false, true},
		{"TestNested", "TestNested/leaf", "pass", "TEST_SET_MISMATCH", 0, 2, 1, false, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			capture := nativeCapture(t, row.name)
			if *capture.ExitCode != row.exit || !slices.Contains(capture.Arguments, "^"+row.name+"$") {
				t.Fatal("original process result or literal filter changed", capture)
			}
			before := bytes.Clone(capture.Stdout)
			plan := nativePlan(t, capture, row.test, row.outcome, row.exit)
			actual, err := Assess(context.Background(), plan, capture)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := json.Marshal(map[string]any{"trial_id": capture.TrialID, "source_revision": capture.SourceRevision,
				"state": row.state, "declared": 1, "observed": row.observed, "matched": row.matched,
				"process_outcome": "MATCH", "argv_captured": true, "expectation_observed": row.accepted,
				"observation_trial_id": capture.TrialID, "observation_source_revision": capture.SourceRevision,
				"failure_observed": row.failure})
			equalJSON(t, actual.Value, want)
			if !bytes.Equal(before, capture.Stdout) {
				t.Fatal("assessment rewrote original observations")
			}
			if row.name == "TestNested" {
				assertAllNamedInventory(t, capture, plan)
			}
			if row.failure {
				capture.Stdout = append(bytes.Clone(capture.Stdout), '{')
				partial, err := Assess(context.Background(), plan, capture)
				var value struct {
					State    string
					Failure  bool `json:"failure_observed"`
					Accepted bool `json:"expectation_observed"`
				}
				if err == nil || json.Unmarshal(partial.Value, &value) != nil ||
					value.State != "INCOMPLETE_OBSERVATION" || !value.Failure || value.Accepted {
					t.Fatal("damaged suffix erased a failure or accepted an incomplete stream", value, err)
				}
			}
		})
	}
}

func assertAllNamedInventory(t *testing.T, capture Capture, leafPlan json.RawMessage) {
	t.Helper()
	var plan map[string]any
	if err := json.Unmarshal(leafPlan, &plan); err != nil {
		t.Fatal(err)
	}
	plan["tests"] = []map[string]string{
		{"test_id": "TestNested/leaf", "outcome": "pass"}, {"test_id": "TestNested", "outcome": "pass"},
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Assess(context.Background(), raw, capture)
	var value struct {
		State                       string
		Declared, Observed, Matched int
		Accepted                    bool `json:"expectation_observed"`
	}
	if err != nil || json.Unmarshal(result.Value, &value) != nil || value.State != "MATCHED_EXPECTATION" ||
		value.Declared != 2 || value.Observed != 2 || value.Matched != 2 || !value.Accepted {
		t.Fatal("explicit all-node inventory did not retain both parent and leaf", value, err)
	}
}
