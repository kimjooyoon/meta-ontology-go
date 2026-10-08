package bodycodegen

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func manyQuadraticCases() string {
	var cases strings.Builder
	for input := int64(1); input <= 10; input++ {
		value := 2*input + 1
		fmt.Fprintf(&cases, "case \"%d\" -> \"%d\"\n", input, input+value*value)
	}
	cases.WriteString("holdout_case \"0\" -> \"1\"\nholdout_case \"11\" -> \"540\"")
	return cases.String()
}

func TestQuadraticHoleManyInputsRetainRequiredAffineFit(t *testing.T) {
	source := quadraticHoleSource("let value = __GOOO_BODY_HOLE_value__; return input + value * value", manyQuadraticCases())
	source = []byte(strings.Replace(string(source), bodySearchHoleQuadraticGrammar, bodySearchHoleQuadraticFitGrammar, 1))
	spec, err := SourceAssembly(context.Background(), "many.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithSourceIRSearch(context.Background(), "many.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodySearch
	if r.TrainingPassed != 10 || r.HoldoutPassed != 2 {
		t.Fatalf("candidate prefix omitted the input-dependent fit: training=%d/%d holdout=%d/%d retained=%d enumerated=%d selected=%q", r.TrainingPassed, r.TrainingTotal, r.HoldoutPassed, r.HoldoutTotal, r.CandidateCount, r.CandidateGeneration.CandidatesEnumerated, r.SelectedExpression)
	}
	if r.CandidateGeneration.CandidatesEnumerated != 108 || r.CandidateGeneration.CandidatesRetained != 16 || r.CandidateGeneration.GrammarComplete {
		t.Fatal("better candidate order erased the truncated grammar boundary")
	}
	legacy := []byte(strings.Replace(string(source), bodySearchHoleQuadraticFitGrammar, bodySearchHoleQuadraticGrammar, 1))
	legacySpec, _ := SourceAssembly(context.Background(), "many.gooo", legacy, "Build")
	before, err := GenerateWithSourceIRSearch(context.Background(), "many.gooo", legacy, "Build", legacySpec, "", "")
	if err != nil || before.Report.BodySearch.TrainingPassed != 1 || before.Report.BodySearch.HoldoutPassed != 0 || before.Report.BodySearch.SelectedExpression != "-3" {
		t.Fatal("v1 candidate order changed", err)
	}
	changed := []byte(strings.Replace(string(source), `"11" -> "540"`, `"12" -> "999"`, 1))
	changedSpec, _ := SourceAssembly(context.Background(), "many.gooo", changed, "Build")
	again, err := GenerateWithSourceIRSearch(context.Background(), "many.gooo", changed, "Build", changedSpec, "", "")
	if err != nil || !reflect.DeepEqual(r.CandidateGeneration, again.Report.BodySearch.CandidateGeneration) {
		t.Fatal("holdout selected the candidate order", err)
	}
}

func TestQuadraticFitOrderRetainsTheOriginalUniverse(t *testing.T) {
	probes := []IRBodyHoleProbe{{Input: 1, Roots: []int64{-3, 3}}, {Input: 2, Roots: []int64{-5, 5}}, {Input: 3, Roots: []int64{-7, 7}}}
	original := quadraticHoleExpressions(probes, []string{"input", "0"})
	ordered := orderQuadraticHoleExpressions(probes, original)
	before, after := slices.Clone(original), slices.Clone(ordered)
	slices.Sort(before)
	slices.Sort(after)
	if !slices.Equal(before, after) || ordered[0] != "input * -2 + -1" || ordered[1] != "input * 2 + 1" {
		t.Fatal("fit ordering changed candidate membership or deterministic prefix", ordered)
	}
	for _, unavailable := range [][]IRBodyHoleProbe{
		{{Input: 1}},
		{{Input: 1, Roots: []int64{3}}, {Input: 1, Roots: []int64{4}}},
	} {
		if !slices.Equal(orderQuadraticHoleExpressions(unavailable, original), original) {
			t.Fatal("absent or contradictory roots manufactured priority")
		}
	}
	if fitsQuadraticRootObservations([]IRBodyHoleProbe{{Input: 1<<63 - 1, Roots: []int64{-1 << 63}}}, 1, 1) {
		t.Fatal("overflow was used as an exact affine observation")
	}
}

func TestQuadraticFitOrderStillScoresTheWholeBody(t *testing.T) {
	source := quadraticHoleSource("let value = __GOOO_BODY_HOLE_value__; return input + value * value * value", "case \"2\" -> \"29\"\ncase \"10\" -> \"37\"")
	source = []byte(strings.Replace(string(source), bodySearchHoleQuadraticGrammar, bodySearchHoleQuadraticFitGrammar, 1))
	spec, _ := SourceAssembly(context.Background(), "cubic.gooo", source, "Build")
	r, err := GenerateWithSourceIRSearch(context.Background(), "cubic.gooo", source, "Build", spec, "", "")
	if err != nil || r.Report.BodySearch.TrainingPassed != 0 || r.Report.BodySearch.TrainingTotal != 2 {
		t.Fatal("root compatibility was treated as whole-body correctness", err)
	}
}
