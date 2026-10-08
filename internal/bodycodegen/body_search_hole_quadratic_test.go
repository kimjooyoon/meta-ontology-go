package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func quadraticHoleSource(body, cases string) []byte {
	return []byte(strings.Replace(string(holeContextSource(body, cases)), bodySearchHoleResidualGrammar, bodySearchHoleQuadraticGrammar, 1))
}

func TestQuadraticHoleRootsUseExactIntegerArithmetic(t *testing.T) {
	for _, a := range []int64{-7, -1, 0, 1, 6} {
		for _, b := range []int64{-5, 0, 3} {
			if a == 0 && b == 0 {
				continue
			}
			for h := int64(-12); h <= 12; h++ {
				roots, status := exactQuadraticHoleRoots(a-b+11, 11, a+b+11, a*h*h+b*h+11)
				if !slices.Contains(roots, h) || !slices.IsSorted(roots) || len(roots) > 2 {
					t.Fatal("exact fitted root missing", a, b, h, roots, status)
				}
			}
		}
	}
	max := int64(1<<63 - 1)
	roots, status := exactQuadraticHoleRoots(max-1, max, max-1, max-9)
	if !reflect.DeepEqual(roots, []int64{-3, 3}) || status != "QUADRATIC_PROPOSAL" {
		t.Fatal("large coefficient arithmetic lost an exact root", roots, status)
	}
	for _, test := range []struct {
		minus, zero, one, expected int64
		status                     string
	}{
		{1, 0, 1, 5, "NO_INTEGER_ROOT_PROPOSAL"},
		{1, 0, 1, -1, "NO_REAL_ROOT_PROPOSAL"},
		{8, 8, 8, 9, "NO_OBSERVED_SENSITIVITY"},
		{-2, 0, 2, 1, "NO_INT64_ROOT_PROPOSAL"},
		{-1 << 63, -1<<63 + 1, -1<<63 + 2, 1, "NO_INT64_ROOT_PROPOSAL"},
	} {
		roots, status := exactQuadraticHoleRoots(test.minus, test.zero, test.one, test.expected)
		if len(roots) != 0 || status != test.status {
			t.Fatal("unavailable integer proposal was invented", roots, status)
		}
	}
}

func TestQuadraticHoleSearchFillsRepeatedValuesAndPreservesLegacy(t *testing.T) {
	ctx := context.Background()
	cases := "case \"2\" -> \"11\"\ncase \"10\" -> \"19\"\nholdout_case \"100\" -> \"109\""
	source := quadraticHoleSource("let factor = __GOOO_BODY_HOLE_value__\nreturn input + factor * factor", cases)
	spec, err := SourceAssembly(ctx, "quadratic.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithSourceIRSearch(ctx, "quadratic.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodySearch
	if r.TrainingPassed != 2 || r.HoldoutPassed != 1 || r.SelectedExpression != "-3" {
		t.Fatal("quadratic proposal failed full-body cases", r.SelectedExpression, r.TrainingPassed, r.HoldoutPassed)
	}
	observed := r.CandidateGeneration.HoleContext
	if observed.Schema != "gooo/integer-hole-context/v2" || observed.EvaluationCalls != 6 || !reflect.DeepEqual(observed.Probes[0].Roots, []int64{-3, 3}) {
		t.Fatal("three-point evidence missing", observed)
	}
	if err := VerifyIRBodySearchProjection(ctx, "quadratic.gooo", source, result); err != nil {
		t.Fatal(err)
	}
	legacy := []byte(strings.Replace(string(source), bodySearchHoleQuadraticGrammar, bodySearchHoleResidualGrammar, 1))
	legacySpec, _ := SourceAssembly(ctx, "quadratic.gooo", legacy, "Build")
	before, err := GenerateWithSourceIRSearch(ctx, "quadratic.gooo", legacy, "Build", legacySpec, "", "")
	if err != nil || before.Report.BodySearch.TrainingPassed != 0 {
		t.Fatal("legacy comparison changed", err)
	}
	raw, _ := json.Marshal(before.Report.BodySearch.CandidateGeneration.HoleContext)
	if strings.Contains(string(raw), "quadratic_roots") || strings.Contains(string(raw), "minus_one_output") {
		t.Fatal("legacy receipt serialization changed")
	}
	observed.Probes[0].Roots[0] = 999
	if err := VerifyIRBodySearchProjection(ctx, "quadratic.gooo", source, result); err == nil {
		t.Fatal("altered root evidence replayed")
	}
}

func TestQuadraticHoleSearchFitsInputDependentValuesAndBranches(t *testing.T) {
	for _, test := range []struct{ body, cases string }{
		{"let value = __GOOO_BODY_HOLE_value__\nreturn input + value * value", "case \"2\" -> \"27\"\ncase \"10\" -> \"451\""},
		{"if input < 0 { return input }; let value = __GOOO_BODY_HOLE_value__; return input + value * value", "case \"-4\" -> \"-4\"\ncase \"2\" -> \"11\""},
		{"let value = __GOOO_BODY_HOLE_value__; return input - value * value", "case \"2\" -> \"-7\"\ncase \"10\" -> \"1\""},
	} {
		source := quadraticHoleSource(test.body, test.cases)
		spec, err := SourceAssembly(context.Background(), "case.gooo", source, "Build")
		if err != nil {
			t.Fatal(err)
		}
		r, err := GenerateWithSourceIRSearch(context.Background(), "case.gooo", source, "Build", spec, "", "")
		if err != nil || r.Report.BodySearch.TrainingPassed != 2 {
			t.Fatal("typed branch/local construction failed", err)
		}
	}
}

func TestQuadraticHoleSearchKeepsNonIntegralAndCubicFitsPartial(t *testing.T) {
	for _, test := range []struct{ body, cases string }{
		{"let value = __GOOO_BODY_HOLE_value__; return input + value * value", "case \"2\" -> \"7\"\ncase \"10\" -> \"15\""},
		{"let value = __GOOO_BODY_HOLE_value__; return input + value * value * value", "case \"2\" -> \"29\"\ncase \"10\" -> \"37\""},
	} {
		source := quadraticHoleSource(test.body, test.cases)
		spec, _ := SourceAssembly(context.Background(), "partial.gooo", source, "Build")
		r, err := GenerateWithSourceIRSearch(context.Background(), "partial.gooo", source, "Build", spec, "", "")
		if err != nil || r.Report.BodySearch.TrainingPassed != 0 || r.Report.BodySearch.TrainingTotal != 2 {
			t.Fatal("probe fit bypassed whole-body scoring", err)
		}
	}
}

func TestQuadraticHoleEvidenceExcludesHoldoutsAndKeepsBounds(t *testing.T) {
	ctx := context.Background()
	source := quadraticHoleSource("let value = __GOOO_BODY_HOLE_value__; return input + value * value", "case \"2\" -> \"11\"\nholdout_case \"100\" -> \"109\"")
	spec, _ := SourceAssembly(ctx, "case.gooo", source, "Build")
	first, err := GenerateWithSourceIRSearch(ctx, "case.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	changed := []byte(strings.Replace(string(source), `"100" -> "109"`, `"101" -> "999"`, 1))
	changedSpec, _ := SourceAssembly(ctx, "case.gooo", changed, "Build")
	second, err := GenerateWithSourceIRSearch(ctx, "case.gooo", changed, "Build", changedSpec, "", "")
	if err != nil || !reflect.DeepEqual(first.Report.BodySearch.CandidateGeneration, second.Report.BodySearch.CandidateGeneration) {
		t.Fatal("holdout changed the probes or candidate list", err)
	}
	plan, _ := sourceIRSearchPlan(spec)
	plan.CandidateGeneration.MaxCandidates = 2
	receipt, err := generateIRBodySearchCandidatesForSource(ctx, "case.gooo", source, "Build", &plan)
	if err != nil || receipt.CandidatesRetained != 2 || receipt.CandidatesOmitted < 1 || receipt.GrammarComplete || receipt.HoleContext.EvaluationCalls != 3 {
		t.Fatal("bounded roots lost coverage or probe counts", err, receipt)
	}
	for _, count := range []int{0, 129} {
		plan.TestCases = make([]IRBodyFillTestCase, count)
		if _, err := generateIRBodySearchCandidatesForSource(ctx, "case.gooo", source, "Build", &plan); err == nil {
			t.Fatal("unbounded quadratic probes accepted", count)
		}
	}
	plan.TestCases = []IRBodyFillTestCase{{Input: 2, Expected: 11}}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := generateIRBodySearchCandidatesForSource(canceled, "case.gooo", source, "Build", &plan); !errors.Is(err, context.Canceled) {
		t.Fatal("quadratic cancellation lost", err)
	}
}

func TestQuadraticHoleUnavailableProbeRetainsFallback(t *testing.T) {
	ctx := context.Background()
	source := quadraticHoleSource("return __GOOO_BODY_HOLE_value__ + 9223372036854775807", "case \"2\" -> \"-9223372036854775807\"")
	spec, _ := SourceAssembly(ctx, "case.gooo", source, "Build")
	r, err := GenerateWithSourceIRSearch(ctx, "case.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	context := r.Report.BodySearch.CandidateGeneration.HoleContext
	p := context.Probes[0]
	if context.EvaluationCalls != 2 || p.MinusOne == nil || p.Zero == nil || p.One != nil || len(p.Roots) != 0 || p.Status != "PROBE_UNAVAILABLE" {
		t.Fatal("missing output was treated as an exact fit", context)
	}
	if r.Report.BodySearch.CandidateGeneration.CandidatesRetained < 2 {
		t.Fatal("unavailable quadratic removed fallback candidates")
	}
}
