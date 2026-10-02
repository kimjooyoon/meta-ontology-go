package main

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func unfixedPathReader(t *testing.T, language string) mapSourceReader {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	name := "../../examples/body-codegen/typed-path-conditional-assignment-plan.json"
	if language == "ko" {
		name = "../../examples/body-codegen/typed-path-conditional-assignment-ko-plan.json"
	}
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var document pathplan.Document
	if json.Unmarshal(raw, &document) != nil {
		t.Fatal("invalid fixture")
	}
	document.Plan.Decisions = document.Plan.Decisions[:2]
	document.MaxAttempts = 4
	document.TestCases = []pathplan.TestCase{{Input: 3, Expected: 999}}
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return mapSourceReader{"fixture.gooo": source, "plan.json": raw,
		"hint.json": []byte(`{"source_sha":"` + strings.Repeat("a", 40) + `","status":"FAIL"}`)}
}

func TestTypedPathUnfixedCLIReconcilesRemainingCoordinatesAndPartialBody(t *testing.T) {
	for _, language := range []string{"en", "ko"} {
		reader := unfixedPathReader(t, language)
		base := []string{"--json", "--path-plan", "plan.json", "--path-model", writeCLIPathFeedbackModel(t),
			"--path-step-attempts", "1", "--path-feedback-rounds", "3", "--path-feedback-ci", "hint.json",
			"--activity", "ConditionalAssign"}
		var legacy bodycodegen.Result
		for _, unfixed := range []bool{false, true} {
			args := append([]string(nil), base...)
			if unfixed {
				args = append(args, "--path-feedback-unfixed")
			}
			args = append(args, "fixture.gooo")
			var out, stderr bytes.Buffer
			if code := runBodyCodegen(args, reader, &out, &stderr); code != exitOK || stderr.Len() != 0 {
				t.Fatalf("CLI %s %t %d: %s %s", language, unfixed, code, out.String(), stderr.String())
			}
			var result bodycodegen.Result
			if json.Unmarshal(out.Bytes(), &result) != nil || result.Report.BodyPaths == nil {
				t.Fatal("missing native path receipt")
			}
			p := result.Report.BodyPaths
			if !p.SourceBaseMatched || p.Search.Evaluated != 4 || p.FunctionalCompleteness != 0 ||
				p.FeedbackUnfixed != unfixed || !result.Report.TypecheckPassed || !result.Report.DeterministicReplay ||
				result.Report.RepositoryWrites != 0 || p.Search.Selection.ExternalCalls != 0 {
				t.Fatal("native partial verification differs")
			}
			if !unfixed {
				legacy = result
				var wire struct {
					Report struct {
						Paths map[string]json.RawMessage `json:"body_paths"`
					} `json:"report"`
				}
				if err := json.Unmarshal(out.Bytes(), &wire); err != nil {
					t.Fatal(err)
				}
				// The opt-in observation stays omitted at its original path; the
				// separately bound search configuration now records false explicitly.
				if p.Search.Selection.ModelCalls != 6 || wire.Report.Paths["feedback_unfixed"] != nil {
					t.Fatal("default receipt changed")
				}
				continue
			}
			if p.Search.Selection.ModelCalls != 5 || result.Source != legacy.Source || result.Report.ActivityID != legacy.Report.ActivityID ||
				!reflect.DeepEqual(p.Search.Attempts, legacy.Report.BodyPaths.Search.Attempts) || len(p.Feedback) != 3 {
				t.Fatal("opt-in changed body, source identity or remaining sequence")
			}
			f := p.Feedback[1]
			if f.ModelCalls != 1 || len(f.FixedCoordinates) != 1 || f.FirstFailure == nil || f.CI == nil ||
				f.CI.Status != "FAIL" || f.CIIsAuthority || f.SHA == "" || !p.Feedback[2].RankingUnnecessary {
				t.Fatal("fixed coordinate proof or original failure lost")
			}
		}
	}
}

func TestTypedPathUnfixedCLIRejectsMissingContextAndDuplicateFlag(t *testing.T) {
	base := []string{"--path-plan", "plan.json", "--path-model", "model.json", "--path-step-attempts", "1"}
	for _, flags := range [][]string{
		{"--path-feedback-unfixed"},
		append(append([]string(nil), base...), "--path-feedback-unfixed"),
		append(append([]string(nil), base...), "--path-feedback-rounds", "1", "--path-feedback-unfixed", "--path-feedback-unfixed"),
	} {
		var out, stderr bytes.Buffer
		args := append(flags, "--activity", "ConditionalAssign", "fixture.gooo")
		if code := runBodyCodegen(args, mapSourceReader{}, &out, &stderr); code != exitUsage {
			t.Fatal("invalid unfixed flag combination read source or loaded a model")
		}
	}
}
