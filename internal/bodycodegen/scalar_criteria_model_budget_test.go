package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func wrongScalarProposalSource(name string, attempts int) []byte {
	return []byte(fmt.Sprintf(`package %s
namespace %s
entity Integer id "urn:%s:integer"
entity Boolean id "urn:%s:boolean"
activity Deduct(Integer) -> Integer id "urn:%s:deduct" computes "return input - 43" assembling {
 choice "direction" operand_order at "0" intent "입력에서 43을 뺀다. Subtract forty-three from the input."
 criteria { predicate "difference" activity "CorrectDifference" input "[8]" input "[29]" holdout_input "[46]" }
 attempts "%d"
}
activity CorrectDifference(Integer,Integer) -> Boolean computes "return input1 == input0 - 43"
`, name, name, name, name, name, attempts))
}

func TestScalarSourceModelKeepsWrongProposalAndBudgetEvidence(t *testing.T) {
	g, err := NewTypedPathGenerator(scalarBridgeSyntheticArtifact(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		name     string
		attempts int
		selected uint16
		status   string
		passed   int
	}{
		{"budgetcomplete", 2, 0, "COMPLETE_FINITE", 2},
		{"budgetlimited", 1, 1, "PARTIAL_FINITE", 0},
	} {
		t.Run(item.name, func(t *testing.T) {
			source := wrongScalarProposalSource(item.name, item.attempts)
			result, err := g.GenerateSourceAssembly(context.Background(), item.name+".gooo", source, "Deduct")
			if err != nil {
				t.Fatal(err)
			}
			r := result.Report.RecordAssembly
			if r == nil || r.ModelCalls != 1 || r.ScalarPrediction == nil || r.ScalarPrediction.ProposedMask != 1 ||
				r.ScalarPrediction.Candidates[0].Probability <= 0.9 || r.SelectedMask != item.selected || r.Status != item.status ||
				len(r.Attempts) != item.attempts || r.Criteria.Passed != item.passed || r.Criteria.Total != 2 ||
				r.Attempts[0].Mask != 1 || r.Attempts[0].Criteria == nil || r.Attempts[0].Criteria.Passed != 0 || r.Attempts[0].Criteria.Failed != 2 {
				t.Fatal("model proposal, source budget or actual condition evidence changed", r)
			}
			if r.CriteriaHoldout.Total != 1 || r.CriteriaHoldout.Passed != item.passed/2 ||
				result.Report.ActivityID != "urn:"+item.name+":deduct" {
				t.Fatal("held condition or source-owned identity changed", r)
			}
			logWrongScalarProposal(t, r)
		})
	}
}

func logWrongScalarProposal(t *testing.T, r *RecordAssemblyReceipt) {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"scope": "fresh wrong synthetic model proposal with source-owned finite budget",
		"source_sha256": r.OriginalSourceSHA256, "proposed_mask": r.ScalarPrediction.ProposedMask,
		"proposed_probability": r.ScalarPrediction.Candidates[0].Probability, "selected_mask": r.SelectedMask,
		"status": r.Status, "attempts": len(r.Attempts), "model_calls": r.ModelCalls,
		"first_proposal_passed": r.Attempts[0].Criteria.Passed, "first_proposal_total": r.Attempts[0].Criteria.Total,
		"selection_passed": r.Criteria.Passed, "selection_total": r.Criteria.Total,
		"held_passed": r.CriteriaHoldout.Passed, "held_total": r.CriteriaHoldout.Total})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("WRONG_SCALAR_PROPOSAL_OBSERVATION", string(raw))
}
