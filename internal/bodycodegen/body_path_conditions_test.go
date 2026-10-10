package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func conditionSource(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../examples/body-codegen/source-condition-cases.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestSourceConditionsConstrainGenerationAndSavedReplay(t *testing.T) {
	ctx := context.Background()
	source := conditionSource(t)
	doc, err := DecodeSourcePathDocument(ctx, "conditions.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"deterministic", "batch", "model", "feedback"} {
		t.Run(mode, func(t *testing.T) {
			var result Result
			var err error
			switch mode {
			case "deterministic":
				result, err = GenerateWithTypedPaths(ctx, "conditions.gooo", source, "Choose", doc, "")
			case "batch":
				result, err = GenerateWithTypedPathBatches(ctx, "conditions.gooo", source, "Choose", doc, "", 1)
			case "model":
				result, err = GenerateWithTypedPaths(ctx, "conditions.gooo", source, "Choose", doc, writeTypedPathContractModel(t, false))
			case "feedback":
				result, err = GenerateWithTypedPathFeedback(ctx, "conditions.gooo", source, "Choose", doc, writeTypedPathContractModel(t, false), 1, 2, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			p := result.Report.BodyPaths
			if p.Conditions == nil || p.Conditions.Passed != 3 || p.Conditions.Declared != 3 ||
				p.Conditions.SelectedTreeSHA256 != p.Conditions.EmittedTreeSHA256 ||
				p.Search.Selection.Choices["comparison"] != "layout_forward" || p.FunctionalCompleteness != 100 {
				t.Fatalf("source conditions missing: %+v", p)
			}
			if err := VerifyTypedPathProjection(ctx, "conditions.gooo", source, doc, result); err != nil {
				t.Fatal(err)
			}
			for _, mutate := range []func(*Result){
				func(r *Result) { r.Report.BodyPaths.Conditions = nil },
				func(r *Result) { r.Report.BodyPaths.Conditions.Results[0].Case.Expected = false },
				func(r *Result) { r.Report.BodyPaths.Search.Selection.Conditions = nil },
				func(r *Result) { r.Report.BodyPaths.Conditions.EmittedTreeSHA256 = digest([]byte("other")) },
			} {
				raw, _ := json.Marshal(result)
				var tampered Result
				if err := json.Unmarshal(raw, &tampered); err != nil {
					t.Fatal(err)
				}
				mutate(&tampered)
				populateCompletenessReceipt(&tampered.Report, "")
				if err := VerifyTypedPathProjection(ctx, "conditions.gooo", source, doc, tampered); err == nil {
					t.Fatal("tampered conditions replayed")
				}
			}
		})
	}
}

func TestSourceConditionsRejectOutputEquivalentCandidateAndChangedDocument(t *testing.T) {
	ctx := context.Background()
	source := conditionSource(t)
	doc, err := DecodeSourcePathDocument(ctx, "conditions.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPaths(ctx, "conditions.gooo", source, "Choose", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	prepared, _ := doc.Prepare()
	wrong := map[string]string{"sign": "layout_forward", "comparison": "layout_reverse", "negative": "layout_forward"}
	program, err := prepared.Compile(wrong)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range doc.TestCases {
		v, err := program.Evaluate(c.Input)
		if err != nil || v.Int != c.Expected {
			t.Fatal("fixture must have output-equivalent wrong predicate", v, err)
		}
	}
	rejected, body, err := RealizeSourcePathCandidate(ctx, "conditions.gooo", source, "Choose", result.Report.BodyPaths, 2)
	var rejection *SourcePathCandidateRejection
	if !errors.As(err, &rejection) || len(body) != 0 || rejected.Stage != "SOURCE_CONDITIONS" || rejected.Conditions.Passed != 1 {
		t.Fatal("caller candidate bypassed declared predicates", rejected, err)
	}
	accepted, body, err := RealizeSourcePathCandidate(ctx, "conditions.gooo", source, "Choose", result.Report.BodyPaths, 1)
	if err != nil || len(body) == 0 || accepted.Conditions.Passed != 3 {
		t.Fatal(accepted, err)
	}
	for _, cases := range [][]pathplan.ConditionCase{nil, {{ChoiceID: "comparison", Input: 0, Expected: true}}} {
		changed := doc
		changed.Plan.ConditionCases = cases
		if _, err := GenerateWithTypedPaths(ctx, "conditions.gooo", source, "Choose", changed, "missing-model.json"); err == nil || !strings.Contains(err.Error(), "source assembling contract") {
			t.Fatal("changed conditions reached model loading", err)
		}
	}
	// A structure comparison must reject the equivalent output even when its
	// two compensating changes leave every finite output case unchanged.
	selected, _ := prepared.Compile(result.Report.BodyPaths.Search.Selection.Choices)
	condition, err := selectedPathConditions(ctx, prepared, result.Report.BodyPaths.Search.Selection.Choices, selected, "Choose", []byte(program.GoSource()))
	if err == nil || condition.Status != "STRUCTURE_MISMATCH" {
		t.Fatal("opposite predicate and branches emitted", condition, err)
	}
}

func TestSourceConditionsImpossibleContractFailsWithoutFallback(t *testing.T) {
	ctx := context.Background()
	source := []byte(strings.Replace(string(conditionSource(t)), `input "0" -> "false"`, `input "0" -> "true"`, 1))
	doc, err := DecodeSourcePathDocument(ctx, "conditions.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = GenerateWithTypedPaths(ctx, "conditions.gooo", source, "Choose", doc, "")
	var failure *BodyPathError
	if !errors.As(err, &failure) || !errors.Is(err, pathplan.ErrNoConditionCandidate) || failure.Receipt.Search.ConditionRejected != 8 {
		t.Fatal("impossible condition escaped or lost attempts", failure, err)
	}
	if failure.Receipt.Conditions == nil || failure.Receipt.Conditions.Status != "UNOBSERVED" || len(failure.Receipt.Conditions.Results) != 0 {
		t.Fatal("fabricated a selected observation")
	}
	// Recipe JSON includes the same typed expectations as source expansion.
	spec, err := SourceAssembly(ctx, "conditions.gooo", source, "Choose")
	if err != nil || !reflect.DeepEqual(sourceAssemblyRecipe(spec).ConditionCases, doc.Plan.ConditionCases) {
		t.Fatal(err)
	}
}

func TestSourceConditionsSkippedNestedIfStaysUnobserved(t *testing.T) {
	source := checkpointFixture(t, "if input < 0 { if input < -2 { return 1 } else { return 2 } } else { return 0 }",
		[]assemblyspec.Choice{{ID: "inner", Kind: "branch_layout", Occurrence: 1, Intent: "Nested branch."}})
	source = []byte(strings.Replace(string(source), `attempts "64"`, `condition_case "inner" input "1" -> "false" attempts "2"`, 1))
	ctx := context.Background()
	doc, err := DecodeSourcePathDocument(ctx, "nested.gooo", source, "Assemble", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = GenerateWithTypedPaths(ctx, "nested.gooo", source, "Assemble", doc, "")
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.Search.ConditionRejected != 2 {
		t.Fatal("skipped condition selected", err)
	}
	for _, attempt := range failure.Receipt.Search.Attempts {
		if len(attempt.Conditions) != 1 || attempt.Conditions[0].Status != "NOT_REACHED" || attempt.Conditions[0].Passed {
			t.Fatal("skipped condition fabricated false", attempt)
		}
	}
}

func TestSourceConditionsUniqueObservationAndNoConditionCompatibility(t *testing.T) {
	ctx := context.Background()
	source := append(conditionSource(t), []byte("\nactivity Expected(Integer) -> Integer computes \"if input < 0 { return 0 - input } else { return input }\"\n")...)
	doc, err := DecodeSourcePathDocument(ctx, "conditions.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPathOptions(ctx, "conditions.gooo", source, "Choose", doc, "", TypedPathOptions{
		Observation: &PathObservationOptions{Inputs: []int64{-11, 11}, MaxCandidates: 8, MaxRounds: 2,
			OracleActivity: "Expected", ReuseProbeOutputs: true, ResolveUniqueCandidate: true}})
	if err != nil {
		t.Fatal(err)
	}
	if !pathResolved(result.Report.BodyPaths) || result.Report.BodyPaths.Conditions.Passed != 3 {
		t.Fatal("unique resolution bypassed conditions")
	}
	if err := VerifyTypedPathProjection(ctx, "conditions.gooo", source, doc, result); err != nil {
		t.Fatal(err)
	}
	var plain []string
	for line := range strings.SplitSeq(string(source), "\n") {
		if !strings.Contains(line, "condition_case") {
			plain = append(plain, line)
		}
	}
	source = []byte(strings.Join(plain, "\n"))
	doc, err = DecodeSourcePathDocument(ctx, "plain.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err = GenerateWithTypedPaths(ctx, "plain.gooo", source, "Choose", doc, "")
	if err != nil || result.Report.BodyPaths.Conditions != nil || len(result.Report.BodyPaths.Search.Selection.Conditions) != 0 {
		t.Fatal("invented conditions for old source", err)
	}
	for _, dimension := range result.Report.CompletenessReceipt.Dimensions {
		if dimension.ID == "typed_path_intermediate_conditions" {
			t.Fatal("invented evidence for absent source conditions")
		}
	}
	if err := VerifyTypedPathProjection(ctx, "plain.gooo", source, doc, result); err != nil {
		t.Fatal(err)
	}
}

func TestSourceConditionsUseReassignedLocalValuesThroughReplay(t *testing.T) {
	source := checkpointFixture(t, "let value = input; value = 0 - value; if value < 0 { return value } else { return 0 - value }",
		[]assemblyspec.Choice{{ID: "sign", Kind: "branch_layout", Occurrence: 0, Intent: "Return a nonnegative value."}})
	source = []byte(strings.Replace(string(source), `case "0" -> "2"`, `case "-9007199254740995" -> "9007199254740995"
 case "0" -> "0"
 case "9007199254740995" -> "9007199254740995"`, 1))
	source = []byte(strings.Replace(string(source), `attempts "64"`, `condition_case "sign" input "-9007199254740995" -> "false"
 condition_case "sign" input "9007199254740995" -> "true"
 attempts "2"`, 1))
	ctx := context.Background()
	doc, err := DecodeSourcePathDocument(ctx, "assigned.gooo", source, "Assemble", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, model := range []string{"", writeTypedPathContractModel(t, false)} {
		result, err := GenerateWithTypedPaths(ctx, "assigned.gooo", source, "Assemble", doc, model)
		if err != nil {
			t.Fatal(err)
		}
		p := result.Report.BodyPaths
		if p.Conditions == nil || p.Conditions.Passed != 2 || p.FunctionalCompleteness != 100 ||
			p.Search.Selection.Choices["sign"] != "layout_reverse" {
			t.Fatal("assignment state was not observed", p)
		}
		if p.Conditions.Results[0].Observation.Value || !p.Conditions.Results[1].Observation.Value {
			t.Fatal("observed original input instead of reassigned local")
		}
		if err := VerifyTypedPathProjection(ctx, "assigned.gooo", source, doc, result); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSourceConditionFailureReachesModelReconsideration(t *testing.T) {
	ctx := context.Background()
	source := strings.ReplaceAll(string(conditionSource(t)),
		`condition_case "comparison" input "-9007199254740995" -> "true"`,
		`condition_case "comparison" input "-9007199254740995" -> "false"`)
	source = strings.ReplaceAll(source,
		`condition_case "comparison" input "9007199254740995" -> "false"`,
		`condition_case "comparison" input "9007199254740995" -> "true"`)
	source = strings.ReplaceAll(source, "음수 입력인지 비교한다. Compare whether input is negative.",
		"양수 입력인지 비교한다. Compare whether input is positive.")
	doc, err := DecodeSourcePathDocument(ctx, "feedback-conditions.gooo", []byte(source), "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	ci := &pathplan.CIHint{SourceSHA: strings.Repeat("a", 40), Status: "FAIL"}
	result, err := GenerateWithTypedPathFeedback(ctx, "feedback-conditions.gooo", []byte(source), "Choose", doc,
		writeTypedPathContractModel(t, false), 1, 2, ci)
	if err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p.Conditions == nil || p.Conditions.Passed != 3 || p.FunctionalCompleteness != 100 || len(p.Feedback) == 0 {
		t.Fatal("condition feedback lost final source validation", p)
	}
	first := p.Feedback[0]
	if first.ConditionRejected != 1 || first.FirstConditionFailure == nil || first.ModelCalls != 3 || len(first.Judgments) != 3 {
		t.Fatal("compiler did not expose the SDK condition feedback", first)
	}
	for _, j := range first.Judgments {
		if !strings.Contains(j.Input, "condition_input=-9007199254740995") ||
			!strings.Contains(j.Input, "condition_expected=false condition_actual=true condition_status=MISMATCH") ||
			!strings.Contains(j.Input, "ci=FAIL") {
			t.Fatal("source-bound model input lost observed condition failure", j.Input)
		}
	}
	if err := VerifyTypedPathProjection(ctx, "feedback-conditions.gooo", []byte(source), doc, result); err != nil {
		t.Fatal(err)
	}
}
