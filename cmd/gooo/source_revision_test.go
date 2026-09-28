package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSourceRevisionCommandsWriteExternalCandidateAndEvaluation(t *testing.T) {
	root := t.TempDir()
	baselinePath := filepath.Join(root, "baseline.gooo")
	inputPath := filepath.Join(root, "input.json")
	baseline := []byte(`package revision
namespace revision
entity Integer id "gooo://revision/entity/integer"
activity Observe(Integer) -> Integer computes "int.add:1"
`)
	if err := os.WriteFile(baselinePath, baseline, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputPath, []byte(`{"value":9223372036854775807}`), 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(baseline)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	candidateDir := filepath.Join(root, "candidate")
	var stdout, stderr bytes.Buffer
	if code := runReviseSource([]string{baselinePath, "--source-digest", digest, "--activity", "Observe", "--expected", "int.add:1", "--replace", "int.add:0", "--reason", "VALUE_INTEGER_OVERFLOW", "--out", candidateDir}, OSFileReader{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("revise code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	evaluationDir := filepath.Join(root, "evaluation")
	stdout.Reset()
	stderr.Reset()
	if code := runEvaluateRevision([]string{baselinePath, filepath.Join(candidateDir, "candidate.gooo"), "--revision", filepath.Join(candidateDir, "revision.json"), "--activity", "Observe", "--input", inputPath, "--out", evaluationDir}, OSFileReader{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("evaluate code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	report, err := os.ReadFile(filepath.Join(evaluationDir, "evaluation.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		State                   string `json:"state"`
		Accepted                bool   `json:"accepted"`
		CounterexampleRecovered bool   `json:"counterexample_recovered"`
	}
	if err := json.Unmarshal(report, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.State != "CLOSED" || decoded.Accepted || !decoded.CounterexampleRecovered {
		t.Fatalf("evaluation = %#v", decoded)
	}
}
