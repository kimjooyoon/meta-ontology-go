package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyrefinement"
)

func TestBodyRefineCLISelectsQuadraticGrammarThroughGooo(t *testing.T) {
	root := "../../examples/quadratic-feedback/"
	args := []string{"body-refine", "--source", root + "source.gooo.fixture", "--activity", "Adjust",
		"--feedback-cases", root + "feedback-cases.json", "--evaluation-cases", root + "evaluation-cases.json",
		"--policy", "../../examples/search-feedback/policy.gooo.fixture", "--search-policy",
		"--max-attempts", "8", "--max-rounds", "4", "--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go"),
		"--out", filepath.Join(t.TempDir(), "run")}
	var out, stderr bytes.Buffer
	if code := run(args, &out, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyrefinement.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Status != "PASS" || result.EvaluationStatus != "PASS" || len(result.Rounds) != 2 {
		t.Fatal("Gooo did not continue through the nonlinear source alternative", err, result.Status, len(result.Rounds))
	}
	if result.Rounds[0].Decision.NextSearchID != "quadratic" || result.Rounds[0].Runtime.FinitePassed != 1 ||
		result.Rounds[1].Runtime.FinitePassed != 3 || result.Evaluation.FinitePassed != 3 || result.Evaluation.ModelCalls != 0 {
		t.Fatal("partial, final or model-free evaluation observation missing")
	}
}
