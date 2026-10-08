package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyrefinement"
)

func TestBodyRefineCLISelectsFitAcrossManyInputs(t *testing.T) {
	root := "../../examples/quadratic-fit-feedback/"
	args := []string{"body-refine", "--source", root + "source.gooo.fixture", "--activity", "Build",
		"--feedback-cases", root + "feedback-cases.json", "--evaluation-cases", root + "evaluation-cases.json",
		"--policy", "../../examples/search-feedback/policy.gooo.fixture", "--search-policy",
		"--max-attempts", "8", "--max-rounds", "4", "--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go"),
		"--out", filepath.Join(t.TempDir(), "run")}
	var out, stderr bytes.Buffer
	if code := run(args, &out, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyrefinement.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Status != "PASS" || len(result.Rounds) != 2 || result.EvaluationStatus != "PASS" {
		t.Fatal("many-case source transition did not complete", err, result.Status)
	}
	if result.Rounds[0].Runtime.FinitePassed != 1 || result.Rounds[0].Decision.NextSearchID != "fit" || result.Rounds[1].Runtime.FinitePassed != 10 || result.Evaluation.FinitePassed != 3 {
		t.Fatal("candidate bound, source transition or finite counts changed")
	}
}

func TestBodyComposeReplaysPublishedQuadraticV1(t *testing.T) {
	raw, err := os.ReadFile("../../docs/research/quadratic-feedback-20261008/refinement.json")
	if err != nil {
		t.Fatal(err)
	}
	var prior bodyrefinement.Result
	if err = json.Unmarshal(raw, &prior); err != nil {
		t.Fatal(err)
	}
	selected, root := prior.Rounds[prior.SelectedRound], t.TempDir()
	for name, value := range map[string]any{"composition.json": selected.Composition, "cases.json": prior.EvaluationCases} {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(root, name), encoded, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err = os.WriteFile(filepath.Join(root, "source.gooo"), []byte(selected.Source), 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	args := []string{"body-compose", "--source", filepath.Join(root, "source.gooo"), "--composition", filepath.Join(root, "composition.json"),
		"--cases", filepath.Join(root, "cases.json"), "--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go")}
	if code := run(args, &out, &stderr); code != exitOK {
		t.Fatal("published v1 record no longer replays", code, stderr.String())
	}
	var replay struct {
		Generated bool `json:"generated_now"`
		Runtime   struct {
			Passed int `json:"finite_passed"`
			Calls  int `json:"model_calls"`
		} `json:"runtime"`
	}
	if err = json.Unmarshal(out.Bytes(), &replay); err != nil || replay.Generated || replay.Runtime.Passed != 3 || replay.Runtime.Calls != 0 {
		t.Fatal("published v1 result or replay path changed", err, replay)
	}
}
