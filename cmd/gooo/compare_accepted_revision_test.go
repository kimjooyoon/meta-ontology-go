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

func TestCompareAcceptedRevisionReportsNextRunImprovement(t *testing.T) {
	root := t.TempDir()
	baselinePath := filepath.Join(root, "baseline.gooo")
	inputPath := filepath.Join(root, "input.json")
	baseline := []byte(`package revision
namespace revision
entity Integer id "revision://entity/integer"
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
		t.Fatalf("revise code=%d stderr=%q", code, stderr.String())
	}
	evaluationDir := filepath.Join(root, "evaluation")
	stdout.Reset()
	stderr.Reset()
	if code := runEvaluateRevision([]string{baselinePath, filepath.Join(candidateDir, "candidate.gooo"), "--revision", filepath.Join(candidateDir, "revision.json"), "--activity", "Observe", "--input", inputPath, "--out", evaluationDir}, OSFileReader{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("evaluate code=%d stderr=%q", code, stderr.String())
	}
	acceptedArgs := []string{baselinePath, filepath.Join(candidateDir, "candidate.gooo"), "--revision", filepath.Join(candidateDir, "revision.json"), "--evaluation", filepath.Join(evaluationDir, "evaluation.json"), "--activity", "Observe", "--input", inputPath, "--accept"}
	stdout.Reset()
	stderr.Reset()
	if code := runAcceptedRevision(acceptedArgs, OSFileReader{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("accepted code=%d stderr=%q", code, stderr.String())
	}
	acceptedPath := filepath.Join(root, "accepted.json")
	if err := os.WriteFile(acceptedPath, stdout.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	compareArgs := []string{baselinePath, filepath.Join(candidateDir, "candidate.gooo"), "--revision", filepath.Join(candidateDir, "revision.json"), "--evaluation", filepath.Join(evaluationDir, "evaluation.json"), "--accepted", acceptedPath, "--activity", "Observe", "--input", inputPath}
	if code := runCompareAcceptedRevision(compareArgs, OSFileReader{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("compare code=%d stderr=%q output=%q", code, stderr.String(), stdout.String())
	}
	var result struct {
		State   string `json:"state"`
		Outcome string `json:"outcome"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.State != "CLOSED" || result.Outcome != "IMPROVED" || result.Reason != "SOURCE_REVISION_IMPROVED_ON_NEXT_RUN" {
		t.Fatalf("comparison = %#v", result)
	}
}
