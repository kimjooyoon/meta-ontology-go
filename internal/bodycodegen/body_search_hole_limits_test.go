package bodycodegen

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestHoleContextUnavailableProbeKeepsLegacyCandidates(t *testing.T) {
	ctx := context.Background()
	source := holeContextSource("return __GOOO_BODY_HOLE_value__ + 9223372036854775807", "case \"2\" -> \"-9223372036854775807\"\ncase \"3\" -> \"-9223372036854775806\"")
	spec, err := SourceAssembly(ctx, "overflow.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithSourceIRSearch(ctx, "overflow.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodySearch
	if r.CandidateGeneration.CandidatesRetained < 2 || r.CandidateGeneration.HoleContext.EvaluationCalls != 2 {
		t.Fatal("failed probes removed the legacy search space")
	}
	for _, probe := range r.CandidateGeneration.HoleContext.Probes {
		if probe.Status != "PROBE_UNAVAILABLE" || probe.Derived != nil || probe.Zero == nil || probe.One != nil {
			t.Fatal("an untypable probe was treated as a valid residual", probe)
		}
	}
	if err := VerifyIRBodySearchProjection(ctx, "overflow.gooo", source, result); err != nil {
		t.Fatal(err)
	}
}

func TestHoleContextBoundsAndCancellation(t *testing.T) {
	source := holeContextSource("return input + __GOOO_BODY_HOLE_value__", "case \"2\" -> \"3\"")
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "bounded.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := sourceIRSearchPlan(spec)
	if err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{0, 129} {
		plan.TestCases = make([]IRBodyFillTestCase, count)
		if _, err := generateIRBodySearchCandidatesForSource(ctx, "bounded.gooo", source, "Build", &plan); err == nil || !strings.Contains(err.Error(), "1..128") {
			t.Fatal("unbounded probes accepted", count, err)
		}
	}
	plan.TestCases = []IRBodyFillTestCase{{Input: 2, Expected: 3, Inputs: []int64{2}}}
	if _, err := generateIRBodySearchCandidatesForSource(ctx, "bounded.gooo", source, "Build", &plan); err == nil {
		t.Fatal("ambiguous scalar input accepted")
	}
	plan.TestCases[0].Inputs = nil
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := generateIRBodySearchCandidatesForSource(canceled, "bounded.gooo", source, "Build", &plan); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation was lost", err)
	}
}

func TestHoleContextTruncationRetainsItsDenominator(t *testing.T) {
	source := holeContextSource("return input + __GOOO_BODY_HOLE_value__", "case \"2\" -> \"13\"\ncase \"10\" -> \"53\"")
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "bounded.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := sourceIRSearchPlan(spec)
	if err != nil {
		t.Fatal(err)
	}
	plan.CandidateGeneration.MaxCandidates = 2
	receipt, err := generateIRBodySearchCandidatesForSource(ctx, "bounded.gooo", source, "Build", &plan)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.GrammarComplete || receipt.CandidatesRetained != 2 || receipt.CandidatesOmitted != receipt.CandidatesEnumerated-2 || receipt.GrammarCoveragePercent >= 100 {
		t.Fatal("truncation lost its denominator", receipt)
	}
	if plan.Candidates[0].Expression != "11" || plan.Candidates[1].Expression != "43" {
		t.Fatal("training-order constants changed")
	}
}
