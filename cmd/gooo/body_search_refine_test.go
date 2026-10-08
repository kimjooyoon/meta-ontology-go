package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyrefinement"
)

func TestBodyRefineCLIExecutesDeclaredSearchAlternatives(t *testing.T) {
	root := "../../examples/search-feedback/"
	args := []string{"body-refine", "--source", root + "source.gooo.fixture", "--activity", "Add",
		"--feedback-cases", root + "feedback-cases.json", "--evaluation-cases", root + "evaluation-cases.json",
		"--policy", root + "policy.gooo.fixture", "--search-policy", "--max-attempts", "8", "--max-rounds", "6",
		"--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go"), "--out", filepath.Join(t.TempDir(), "run")}
	var out, stderr bytes.Buffer
	if code := run(args, &out, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyrefinement.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || !result.SearchPolicy || result.Status != "PASS" || result.EvaluationStatus != "PASS" || len(result.Rounds) != 4 {
		t.Fatal("CLI did not expose the actual grammar revision loop", err, result.Status)
	}
}
