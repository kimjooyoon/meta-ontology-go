package bodycodegen

import (
	"context"
	"encoding/json"
	"testing"
)

const feeSharedInputScalarSource = `package feeshared
namespace feeshared
entity Integer id "urn:feeshared:integer"
entity Boolean id "urn:feeshared:boolean"
activity Fee(Integer) -> Integer id "urn:feeshared:fee" computes "if input < 0 { return 0 } else { if input > 45 { return input*2 } else { return 90 } }" assembling {
 choice "cap" branch_layout at "1" intent "Return at most 90 after multiplying the nonnegative input by two."
 choice "negative" branch_layout at "0" intent "Keep negative input on the zero-fee branch."
 criteria {
  predicate "nonnegative" activity "Nonnegative"
  predicate "required" activity "RequiredFee"
  input "[-3]" input "[0]" input "[21]" input "[99]"
  holdout_input "[-7]" holdout_input "[30]" holdout_input "[45]" holdout_input "[46]"
 }
 attempts "2"
}
activity Nonnegative(Integer,Integer) -> Boolean id "urn:feeshared:nonnegative" computes "return input1 >= 0"
activity RequiredFee(Integer,Integer) -> Boolean id "urn:feeshared:required" computes "if input0 < 0 { return input1 == 0 } else { if input0 > 45 { return input1 == 90 } else { return input1 == input0*2 } }"
`

func TestScalarSourceModelAssemblesFeeBranchesWithSharedPredicateInputs(t *testing.T) {
	g, err := NewTypedPathGenerator(scalarBridgeSyntheticArtifact(t))
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "fee-shared-inputs.gooo", []byte(feeSharedInputScalarSource), "Fee")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r == nil || r.ScalarPrediction == nil || r.ModelCalls != 1 || r.SelectedMask != 1 ||
		r.ScalarPrediction.CandidateCount != 4 || r.ScalarPrediction.ProposedMask != 3 || len(r.Attempts) != 2 ||
		r.Status != "COMPLETE_FINITE" || r.Criteria.Passed != 8 || r.Criteria.Total != 8 ||
		r.CriteriaHoldout.Passed != 8 || r.CriteriaHoldout.Total != 8 {
		t.Fatal("bounded fee branches did not satisfy their fixed criteria", r)
	}
	if r.Attempts[0].Criteria.Passed != 4 || r.Attempts[0].Criteria.Total != 8 ||
		r.Attempts[0].Mask != 3 || r.Attempts[1].Mask != 1 || result.Report.ActivityID != "urn:feeshared:fee" {
		t.Fatal("first proposal outcomes or source identity changed", r)
	}
	raw, err := json.Marshal(map[string]any{"source_sha256": r.OriginalSourceSHA256,
		"choices": len(r.Choices), "candidate_count": r.ScalarPrediction.CandidateCount,
		"proposed_mask": r.ScalarPrediction.ProposedMask, "selected_mask": r.SelectedMask,
		"first_proposal_passed": r.Attempts[0].Criteria.Passed, "first_proposal_total": r.Attempts[0].Criteria.Total,
		"selection_passed": r.Criteria.Passed, "selection_total": r.Criteria.Total,
		"held_passed": r.CriteriaHoldout.Passed, "held_total": r.CriteriaHoldout.Total, "model_calls": r.ModelCalls})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("FEE_SHARED_INPUT_SCALAR_OBSERVATION", string(raw))
}
