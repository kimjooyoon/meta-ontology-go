package bodypathstream

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilesTimingBindsOriginalFilesAndCurrentPhases(t *testing.T) {
	args, out := fileArgs(t, true, `{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":2,"expected":5},{"input":-4,"expected":11}]}`)
	args = append(args, "--timing")
	var stdout, stderr bytes.Buffer
	if code := RunFilesCommand(context.Background(), "files", args, &stdout, &stderr); code != 0 {
		t.Fatal(code, stderr.String())
	}
	results := decodeResults(t, stdout.Bytes())
	for i, result := range results {
		name := fmt.Sprintf("run-%d-timing.json", i+1)
		raw, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		var timing fileTiming
		if err = json.Unmarshal(raw, &timing); err != nil {
			t.Fatal(err)
		}
		if timing.Schema != "gooo/body-path-file-timing/v1" || timing.Status != result.Status || timing.Wall.Status != "OBSERVED" {
			t.Fatal(timing)
		}
		if err = verifyFileTiming(out, timing); err != nil {
			t.Fatal(err)
		}
		for _, mutate := range []func(*fileTiming){
			func(t *fileTiming) { t.Wall.Phases[0].StartNS = -1 },
			func(t *fileTiming) { t.Wall.Phases[0].Outcome = "invented" },
			func(t *fileTiming) { t.Wall.UnassignedNS++ },
			func(t *fileTiming) { t.ResponseNS = t.Wall.CaptureNS + 1 },
			func(t *fileTiming) { t.Files["../outside"] = "sha256:x" },
		} {
			var changed fileTiming
			if err := json.Unmarshal(raw, &changed); err != nil {
				t.Fatal(err)
			}
			mutate(&changed)
			if verifyFileTiming(out, changed) == nil {
				t.Fatal("invalid timing accepted", changed)
			}
		}
		seen := map[string]bool{}
		var sum int64
		for _, p := range timing.Wall.Phases {
			seen[p.Name] = true
			sum += p.EndNS - p.StartNS
		}
		if sum+timing.Wall.UnassignedNS != timing.Wall.CaptureNS || timing.ResponseNS <= 0 ||
			!seen["generation"] || !seen["source_replay"] || !seen["go_tool_hash"] || !seen["native_run_1"] ||
			!seen["native_run_2"] || !seen["artifact_save"] || !seen["executable_hash_1"] || !seen["executable_hash_2"] {
			t.Fatal("phase accounting incomplete", timing)
		}
		if result.Execution.CompletenessReceipt.ProfileID != "gooo/typed-path-runtime-v3" {
			t.Fatal("diagnostic changed runtime profile")
		}
		response, err := os.ReadFile(filepath.Join(out, fmt.Sprintf("run-%d-response.json", i+1)))
		if err != nil {
			t.Fatal(err)
		}
		if timing.Files[fmt.Sprintf("run-%d-response.json", i+1)] != fmt.Sprintf("sha256:%x", sha256.Sum256(response)) {
			t.Fatal("wrong response binding")
		}
		timing.Files[fmt.Sprintf("run-%d-response.json", i+1)] = "sha256:" + strings.Repeat("0", 64)
		if verifyFileTiming(out, timing) == nil {
			t.Fatal("changed file binding accepted")
		}
	}
	if _, err := os.Stat(filepath.Join(out, "timing-summary.json")); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := RunFilesCommand(context.Background(), "files", []string{"--verify-timing", "--out", out},
		&stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "PASS") {
		t.Fatal("saved timing verification failed", code, stderr.String())
	}
	rows := fileSummary(t, out)
	if len(rows) != 2 || rows[0].Passed != 2 || !rows[1].ArtifactReused {
		t.Fatal("old summary changed", rows)
	}
}

func TestTimingReaderRejectsAliasesDuplicatesAndBadIntervals(t *testing.T) {
	for _, raw := range []string{`{"Schema":"x"}`, `{"schema":"x","schema":"y"}`,
		`{"wall":{"Status":"OBSERVED"}}`, `{"schema":"x"} {}`} {
		var timing fileTiming
		if decodeTiming([]byte(raw), &timing) == nil {
			t.Fatal("noncanonical timing accepted", raw)
		}
	}
	var timing fileTiming
	if err := decodeTiming([]byte(`{"files":{"run-1-response.json":"sha256:x"},"wall":{"phases":[]}}`), &timing); err != nil {
		t.Fatal("literal filename rejected", err)
	}
}

func TestTimingPreservesFailedToolLookupWithoutInventingNativeWork(t *testing.T) {
	args, out := fileArgs(t, true, `{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":2,"expected":5}]}`)
	args = append(args, "--timing", "--go-bin", filepath.Join(t.TempDir(), "missing-go"))
	var stdout, stderr bytes.Buffer
	if code := RunFilesCommand(context.Background(), "files", args, &stdout, &stderr); code != 1 {
		t.Fatal(code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "native unobserved") {
		t.Fatal("unstarted native duration displayed as zero", stderr.String())
	}
	raw, err := os.ReadFile(filepath.Join(out, "run-1-timing.json"))
	if err != nil {
		t.Fatal(err)
	}
	var timing fileTiming
	if err := json.Unmarshal(raw, &timing); err != nil {
		t.Fatal(err)
	}
	if timing.Status != "execution_failed" || timing.Wall.Status != "OBSERVED" {
		t.Fatal(timing)
	}
	failed := false
	for _, p := range timing.Wall.Phases {
		if p.Name == "runtime_suite_prepare" && p.Outcome == "failed" {
			failed = true
		}
		if strings.HasPrefix(p.Name, "native_run_") {
			t.Fatal("unstarted native work recorded")
		}
	}
	if !failed {
		t.Fatal("failed stage lost")
	}
	if err := verifyTimingDirectory(context.Background(), out); err != nil {
		t.Fatal(err)
	}
}

func TestPhaseDurationRequiresEveryGroupMember(t *testing.T) {
	for _, tc := range []struct {
		phases map[string]float64
		want   string
	}{{nil, "unobserved"}, {map[string]float64{"a": 1}, "unobserved"},
		{map[string]float64{"a": 0, "b": 0}, "0.000ms"},
		{map[string]float64{"a": 1.25, "b": 2.5}, "3.750ms"}} {
		if got := phaseDuration(tc.phases, "a", "b"); got != tc.want {
			t.Fatal(got, tc.want)
		}
	}
}

func TestFilesTimingKeepsRejectedRequestsAndFailureStages(t *testing.T) {
	args, out := fileArgs(t, true, `{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":2}]}`)
	args = append(args, "--timing")
	var stdout, stderr bytes.Buffer
	if code := RunFilesCommand(context.Background(), "files", args, &stdout, &stderr); code != 1 {
		t.Fatal(code, stderr.String())
	}
	raw, err := os.ReadFile(filepath.Join(out, "run-1-timing.json"))
	if err != nil {
		t.Fatal(err)
	}
	var timing fileTiming
	if err = json.Unmarshal(raw, &timing); err != nil {
		t.Fatal(err)
	}
	if timing.Status != "rejected" || timing.Wall.Status != "OBSERVED" {
		t.Fatal(timing)
	}
	for _, p := range timing.Wall.Phases {
		if p.Name == "generation" || strings.HasPrefix(p.Name, "native_run_") {
			t.Fatal("unexecuted phase invented", p)
		}
	}
}
