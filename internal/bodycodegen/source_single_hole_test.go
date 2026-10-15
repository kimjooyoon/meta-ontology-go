package bodycodegen

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

func TestSourceSingleHoleTextLoop(t *testing.T) {
	source, activity := singleHoleByteFixture(t)
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "single-byte-fill.gooo", source,
		"IndexByte", &activity.Assembly.Spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertSingleHoleTextFill(t, result)
	if result.Report.BodyFill.LocalModelPredictions != 0 || strings.Contains(result.GoooSource, "source_fill") ||
		strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") {
		t.Fatal("deterministic source fill retained pending choices or inferred", result.Report.BodyFill)
	}
}

// The chooser is simulated. This regression does not measure neural accuracy.
func TestSourceSingleHoleTextLoopRetainsWrongProposal(t *testing.T) {
	source, activity := singleHoleByteFixture(t)
	plan, _, err := sourceIRBodyFillPlan("single-byte-fill.gooo", source, "IndexByte", &activity.Assembly.Spec)
	if err != nil {
		t.Fatal(err)
	}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationLessThan}
	result, err := generateWithIRBodyFillOptions(context.Background(), "single-byte-fill.gooo", source,
		"IndexByte", plan, "", "", IRBodyFillOptions{}, provider)
	if err != nil {
		t.Fatal(err)
	}
	assertSingleHoleTextFill(t, result)
	r := result.Report.BodyFill
	if provider.calls != 1 || r.LocalModelPredictions != 1 || r.ProposedCandidateID != "lower" ||
		r.ProposedAccuracyPct != 100.0/6 || r.CandidateScores[0].TestCasesPassed != 1 ||
		r.CandidateScores[0].TestCasesTotal != 6 {
		t.Fatal("simulated wrong proposal or its finite score was lost", r)
	}
}

func assertSingleHoleTextFill(t *testing.T, result Result) {
	t.Helper()
	r := result.Report.BodyFill
	if r == nil || r.SelectedCandidateID != "equal" || r.TestCasesPassed != 6 || r.TestCasesTotal != 6 ||
		r.HoldoutCasesPassed != 4 || r.HoldoutCasesTotal != 4 || r.ExternalProviderCalls != 0 ||
		!result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatal("single-hole body did not retain checked finite outcomes", r)
	}
}
