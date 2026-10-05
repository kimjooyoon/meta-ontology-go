package bodycodegen

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestSourceFillAssignmentEnumerationIsStableAndReportsEachBound(t *testing.T) {
	plan := &assemblyspec.FillPlan{
		Intent: "Compose a small integer result.",
		Holes:  []assemblyspec.FillHole{{ID: "left"}, {ID: "right"}},
		Generation: &assemblyspec.FillGeneration{
			Grammar: bodySearchIntegerAffineGrammar, MaxExpressions: 2, MaxCandidates: 16,
		},
	}
	cases := []assemblyspec.Case{{Input: 0, Expected: 1}, {Input: 2, Expected: 3}}
	first, firstReceipt, err := generateSourceFillCandidates(plan, cases)
	if err != nil {
		t.Fatal(err)
	}
	second, secondReceipt, err := generateSourceFillCandidates(plan, cases)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) || firstReceipt.CandidateSetSHA256 != secondReceipt.CandidateSetSHA256 {
		t.Fatalf("source-derived candidate enumeration is not stable: first=%#v second=%#v", first, second)
	}
	if firstReceipt.GrammarComplete || firstReceipt.GrammarCoveragePercent >= 100 ||
		firstReceipt.AssignmentSpaceSize != 4 || firstReceipt.AssignmentsRetained != 4 ||
		firstReceipt.AssignmentsOmitted != 0 || firstReceipt.AssignmentCoveragePercent != 100 {
		t.Fatalf("grammar truncation and assignment completeness were conflated: %+v", firstReceipt)
	}
}
