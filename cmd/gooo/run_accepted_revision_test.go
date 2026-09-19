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

func TestRunAcceptedRevisionRequiresExplicitAcceptanceAndReexecutesCandidate(t *testing.T) {
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
	args := []string{baselinePath, filepath.Join(candidateDir, "candidate.gooo"), "--revision", filepath.Join(candidateDir, "revision.json"), "--evaluation", filepath.Join(evaluationDir, "evaluation.json"), "--activity", "Observe", "--input", inputPath}
	stdout.Reset()
	stderr.Reset()
	if code := runAcceptedRevision(args, OSFileReader{}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("missing accept code=%d stderr=%q", code, stderr.String())
	}
	args = append(args, "--accept")
	stdout.Reset()
	stderr.Reset()
	if code := runAcceptedRevision(args, OSFileReader{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("accepted code=%d stderr=%q", code, stderr.String())
	}
	var result struct {
		Decision          string `json:"decision"`
		ExecutionAllowed  bool   `json:"execution_allowed"`
		RepositoryWrites  int    `json:"repository_writes"`
		Execution         struct {
			Results map[string]struct {
				Value int64 `json:"value"`
			} `json:"results"`
		} `json:"execution"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Decision != "PASS" || result.ExecutionAllowed || result.RepositoryWrites != 0 || result.Execution.Results["Observe"].Value != 9223372036854775807 {
		t.Fatalf("result = %#v", result)
	}
}
