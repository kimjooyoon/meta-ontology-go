package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

var goErrorGuardNativePaths = []string{
	"TestRunSourcePackageOutputFailureIsNotSuccess/human",
	"TestRunSourcePackageOutputFailureIsNotSuccess/json",
	"TestRunSourcePrintsHumanPackageSummary",
	"TestRunSourceJSONPackageReplayIsByteStable",
}

func TestGoErrorGuardGeneratedCandidateUsesFrozenNativeOracle(t *testing.T) {
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(directory, "..", "..", ".."))
	oraclePath := filepath.Join(root, "cmd", "gooo", "run_package_source_test.go")
	oracle, err := os.ReadFile(oraclePath)
	if err != nil {
		t.Fatal(err)
	}
	// Freeze the unchanged repository oracle before candidate generation.
	// A build overlay does not change runtime filesystem reads.
	temp := t.TempDir()
	writeGoGuardNativeFile(t, filepath.Join(temp, "oracle.go"), oracle)
	program := goErrorGuardFixture()
	proposal, err := ProposeGoErrorGuard("guard.gooo", program, goErrorGuardOriginal)
	if err != nil {
		t.Fatal(err)
	}
	writeGoGuardNativeFile(t, filepath.Join(temp, "guard.gooo"), program)
	before, beforeCode := runGoGuardNativeOverlay(t, root, temp, "before", goErrorGuardOriginal)
	after, afterCode := runGoGuardNativeOverlay(t, root, temp, "after", []byte(proposal.CandidateSource))
	if beforeCode != 1 || afterCode != 0 {
		t.Fatalf("native process exits before=%d after=%d", beforeCode, afterCode)
	}
	for index, path := range goErrorGuardNativePaths {
		wantBefore := "pass"
		if index == 0 {
			wantBefore = "fail"
		}
		if len(before[path]) != 1 || before[path][0] != wantBefore || len(after[path]) != 1 || after[path][0] != "pass" {
			t.Fatalf("frozen leaf %s: before=%v after=%v", path, before[path], after[path])
		}
	}
	currentOracle, err := os.ReadFile(oraclePath)
	if err != nil || !bytes.Equal(oracle, currentOracle) {
		t.Fatal("native trial changed the input oracle")
	}
	t.Logf("gooo guard witness: program=%s ir=%s activity=%s original=%s candidate=%s oracle=%s paths=4 before_pass=3 before_fail=1 after_pass=4 after_fail=0 adoption=UNKNOWN utility=UNKNOWN performance=UNKNOWN",
		proposal.ProgramDigest, proposal.SemanticDigest, proposal.ActivityID,
		proposal.SourceDigest, proposal.CandidateDigest, DigestBytes(oracle))
}

func runGoGuardNativeOverlay(t *testing.T, root, temp, trial string, source []byte) (map[string][]string, int) {
	t.Helper()
	backing := filepath.Join(temp, trial+".go")
	writeGoGuardNativeFile(t, backing, source)
	overlay, err := json.Marshal(map[string]map[string]string{"Replace": {
		filepath.Join(root, "cmd", "gooo", "run_package_source.go"):      backing,
		filepath.Join(root, "cmd", "gooo", "run_package_source_test.go"): filepath.Join(temp, "oracle.go"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(temp, trial+"-overlay.json")
	writeGoGuardNativeFile(t, overlayPath, overlay)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pattern := "^(TestRunSourcePackageOutputFailureIsNotSuccess|TestRunSourcePrintsHumanPackageSummary|TestRunSourceJSONPackageReplayIsByteStable)$"
	command := exec.CommandContext(ctx, "go", "test", "-json", "-count=1", "-mod=readonly", "-overlay", overlayPath, "-run", pattern, "./cmd/gooo")
	command.Dir = root
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err = command.Run()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || ctx.Err() != nil {
			t.Fatalf("native %s could not finish: %v stderr=%s", trial, err, stderr.String())
		}
		code = exit.ExitCode()
	}
	writeGoGuardNativeFile(t, filepath.Join(temp, trial+"-events.jsonl"), stdout.Bytes())
	writeGoGuardNativeFile(t, filepath.Join(temp, trial+"-stderr.txt"), stderr.Bytes())
	t.Logf("native %s command=%q source=%s stdout_digest=%s stderr_digest=%s exit=%d\nstdout:\n%s\nstderr:\n%s",
		trial, command.Args, DigestBytes(source), DigestBytes(stdout.Bytes()), DigestBytes(stderr.Bytes()), code, stdout.String(), stderr.String())
	return collectGoGuardNativeEvents(t, stdout.Bytes()), code
}

func collectGoGuardNativeEvents(t *testing.T, raw []byte) map[string][]string {
	t.Helper()
	results := make(map[string][]string)
	decoder := json.NewDecoder(bytes.NewReader(raw))
	for {
		var event struct {
			Action, Test string
		}
		err := decoder.Decode(&event)
		if err == io.EOF {
			return results
		}
		if err != nil {
			t.Fatalf("native event is not JSON: %v", err)
		}
		if event.Action != "pass" && event.Action != "fail" {
			continue
		}
		for _, path := range goErrorGuardNativePaths {
			if path == event.Test {
				results[path] = append(results[path], event.Action)
			}
		}
	}
}

func writeGoGuardNativeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
