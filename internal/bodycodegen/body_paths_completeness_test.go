package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func consumePathCompleteness(t *testing.T, receipt *CompletenessReceipt) {
	t.Helper()
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := completeness.Decode(raw)
	if err != nil || decoded.ProfileID != typedPathCompletenessProfile || decoded.AggregateCompletenessScore != nil {
		t.Fatalf("shared consumer: %v, %+v", err, decoded)
	}
	for _, id := range []string{"execution_boundary", "full_domain_semantics", "reverse_observation_coverage"} {
		if completenessDimensionByID(t, decoded, id).Status != "UNKNOWN" {
			t.Fatalf("finite AST scores claimed %s", id)
		}
	}
}

func TestTypedPathCompletenessFiniteScoresAreNotLoweringScores(t *testing.T) {
	for _, score := range []int{3, 2, 0} {
		t.Run(string(rune('0'+score)), func(t *testing.T) {
			source, document := typedPathFixture(t)
			for i := score; i < len(document.TestCases); i++ {
				document.TestCases[i].Expected = 999999
			}
			result, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, "")
			if err != nil {
				t.Fatal(err)
			}
			consumePathCompleteness(t, result.Report.CompletenessReceipt)
			d := completenessDimensionByID(t, result.Report.CompletenessReceipt, "typed_path_finite_accuracy")
			want := "PROGRESS"
			if score == 3 {
				want = "PASS"
			}
			if d.Numerator != score || d.Denominator != 3 || d.Status != want || result.Report.CompletenessPercent != 100 {
				t.Fatalf("finite/lowering distinction: %+v, %v", d, result.Report.CompletenessPercent)
			}
			for _, id := range []string{"typed_path_source_binding", "typed_path_provider_accounting", "external_network_boundary"} {
				if completenessDimensionByID(t, result.Report.CompletenessReceipt, id).Status != "PASS" {
					t.Fatalf("lost observed axis: %s", id)
				}
			}
			// A clean compiler pin isolates the core decision in this unit fixture.
			report := result.Report
			report.CompilerSourceSHA = strings.Repeat("a", 40)
			r := buildCompletenessReceipt(report, "")
			if (r.Decision == "PASS_WITHIN_DECLARED_FIXTURE_SCOPE") != (score == 3) {
				t.Fatalf("core decision hid finite score: %s", r.Decision)
			}
		})
	}
}

func TestTypedPathCompletenessPlanBindsIntentAndSearchControls(t *testing.T) {
	source, document := typedPathFixture(t)
	first, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, "")
	if err != nil {
		t.Fatal(err)
	}
	document.Plan.Decisions[0].Intent = "다른 표현으로 같은 경로를 설명 / alternate English intent"
	second, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, "")
	if err != nil || first.Source != second.Source || first.Report.PlanSHA256 == second.Report.PlanSHA256 {
		t.Fatalf("same emission erased distinct intent plan: %v", err)
	}
	third, err := GenerateWithTypedPathBatches(context.Background(), "fixture.gooo", source, "Combined", document, "", 1)
	if err != nil || second.Source != third.Source || second.Report.PlanSHA256 == third.Report.PlanSHA256 {
		t.Fatalf("same emission erased different search controls: %v", err)
	}
}

func TestTypedPathCompletenessSeparatesModelFromEmissionProvider(t *testing.T) {
	source, document := typedPathFixture(t)
	result, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document,
		writeTypedPathContractModel(t, false))
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.CompletenessReceipt
	consumePathCompleteness(t, receipt)
	scope := receipt.Scope["typed_path"].(map[string]any)
	if scope["decision_provider"] != "tiny_go" || scope["local_model_predictions"] != 3 ||
		scope["external_provider_calls"] != 0 || scope["provider_counts_known"] != true ||
		receipt.Scope["emission_decision_provider"] != "deterministic" || receipt.Scope["laya_provider"] != "" {
		t.Fatalf("provider stages conflated: %+v", receipt.Scope)
	}
	if completenessDimensionByID(t, receipt, "typed_path_provider_accounting").Status != "PASS" {
		t.Fatal("actual prediction count is unbound")
	}
}

func TestTypedPathFailureKeepsKnownZeroCallsAndUnknownScore(t *testing.T) {
	source, document := typedPathFixture(t)
	source = []byte(strings.Replace(string(source), "input + 2", "input + 99", 1))
	_, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, "missing-model.json")
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.SearchStarted {
		t.Fatalf("expected pre-model failure: %v", err)
	}
	r := PathFailureCompletenessReceipt("Combined", source, err.Error(), failure.Receipt)
	consumePathCompleteness(t, r)
	finite := completenessDimensionByID(t, r, "typed_path_finite_accuracy")
	if r.Decision != "FAIL_CLOSED" || finite.Status != "UNKNOWN" || finite.Numerator != 0 || finite.Denominator != 3 ||
		completenessDimensionByID(t, r, "typed_path_provider_accounting").Status != "PASS" {
		t.Fatalf("pre-model failure lost observations: %+v", r)
	}
	// Once entered, an absent SDK result cannot be inferred to mean zero calls.
	failure.Receipt.SearchStarted = true
	r = PathFailureCompletenessReceipt("Combined", source, err.Error(), failure.Receipt)
	consumePathCompleteness(t, r)
	if completenessDimensionByID(t, r, "typed_path_provider_accounting").Status != "UNKNOWN" ||
		completenessDimensionByID(t, r, "external_network_boundary").Status != "UNKNOWN" {
		t.Fatal("missing search result was rewritten as known zero")
	}
}

func TestTypedPathCompletenessRejectsContradictoryFiniteEvidence(t *testing.T) {
	for _, mutation := range []string{"actual", "suite", "count", "source", "config", "prediction-count", "network"} {
		t.Run(mutation, func(t *testing.T) {
			source, document := typedPathFixture(t)
			result, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, "")
			if err != nil {
				t.Fatal(err)
			}
			p := result.Report.BodyPaths
			switch mutation {
			case "actual":
				p.NativeCases[0].Actual++
			case "suite":
				p.TestSuiteSHA256 = digest([]byte("different suite"))
			case "count":
				p.Search.SelectedTrainingPassed = 0
			case "source":
				p.SelectedSourceSHA256 = digest([]byte("different source"))
			case "config":
				p.SearchConfig = json.RawMessage(`{"step_attempts":63}`)
			case "prediction-count":
				p.Search.Selection.ModelCalls = -1
			case "network":
				p.Search.Selection.ExternalCalls = 1
			}
			r := buildCompletenessReceipt(result.Report, "")
			consumePathCompleteness(t, r)
			if r.Decision != "FAIL_CLOSED" {
				t.Fatal("contradictory observation accepted")
			}
		})
	}
}
