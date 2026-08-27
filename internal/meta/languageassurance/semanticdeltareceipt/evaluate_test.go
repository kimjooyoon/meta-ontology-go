package semanticdeltareceipt

import (
	"reflect"
	"testing"
)

const candidateSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestSuiteHasFixedDenominatorAndDeterministicReplay(t *testing.T) {
	first, second := RunSuite(candidateSHA), RunSuite(candidateSHA)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("semantic delta suite replay differs")
	}
	if first.Decision != DecisionFixedPoint || first.Resolution != ResolutionExact || first.Summary.CasesTotal != 3 ||
		first.Summary.CasesPassed != 3 || first.CoverageBPS != 10000 || first.Summary.TextualChanges != 3 ||
		first.Summary.SemanticPreserved != 1 || first.Summary.SemanticChanged != 1 || first.Summary.Indeterminate != 1 {
		t.Fatalf("suite=%+v", first)
	}
}

func TestTextualDeltaCanPreserveMeaning(t *testing.T) {
	report := Evaluate(CaseInput("equivalent", candidateSHA))
	if !report.IndependentVerdict.Passed || report.Receipt.Decision != DecisionFixedPoint ||
		report.Receipt.Classification != ClassPreserved || !report.Receipt.TextualDelta.Changed ||
		!structuralIsEmpty(report.Receipt.StructuralDelta) || !claimIsEmpty(report.Receipt.SemanticClaimDelta) {
		t.Fatalf("report=%+v", report)
	}
}

func TestSmallTextualChangeCanBreakMeaning(t *testing.T) {
	report := Evaluate(CaseInput("semantic-change", candidateSHA))
	var generated ClaimTransition
	for _, transition := range report.Receipt.ClaimTransitions {
		if transition.FromObject == "gooo://semantic-delta/entity/payment" {
			generated = transition
		}
	}
	if !report.IndependentVerdict.Passed || report.Receipt.Decision != DecisionDelta ||
		report.Receipt.Classification != ClassChanged || len(report.Receipt.StructuralDelta.AddedNodes) != 1 ||
		len(report.Receipt.SemanticClaimDelta.Changed) != 1 || generated.ToObject != "gooo://semantic-delta/entity/reversal" {
		t.Fatalf("report=%+v", report)
	}
}

func TestUnparseableSourceIsIndeterminateAndFailClosed(t *testing.T) {
	report := Evaluate(CaseInput("indeterminate", candidateSHA))
	if !report.IndependentVerdict.Passed || report.IndependentVerdict.Decision != DecisionFailClosed ||
		report.IndependentVerdict.Resolution != ResolutionUnknown || report.Receipt.Classification != ClassIndeterminate ||
		report.Receipt.StructuralDelta.Status != "UNKNOWN" || report.Receipt.SemanticClaimDelta.Status != "UNKNOWN" ||
		report.RepositoryWrites != 0 {
		t.Fatalf("report=%+v", report)
	}
}

func TestIndependentAdjudicatorRejectsTamperedReceipt(t *testing.T) {
	input := CaseInput("equivalent", candidateSHA)
	receipt, err := Produce(input)
	if err != nil {
		t.Fatal(err)
	}
	receipt.TextualDelta.ChangedBytes++
	verdict := Adjudicate(input, receipt)
	if verdict.Passed || verdict.Reason != ReasonReceipt || verdict.Resolution != ResolutionInvariant {
		t.Fatalf("verdict=%+v", verdict)
	}
}

func structuralIsEmpty(delta StructuralDelta) bool {
	return delta.Status == "KNOWN" && len(delta.AddedNodes) == 0 && len(delta.RemovedNodes) == 0 &&
		len(delta.AddedFacts) == 0 && len(delta.RemovedFacts) == 0
}

func claimIsEmpty(delta ClaimDelta) bool {
	return delta.Status == "KNOWN" && len(delta.Added) == 0 && len(delta.Removed) == 0 && len(delta.Changed) == 0
}
