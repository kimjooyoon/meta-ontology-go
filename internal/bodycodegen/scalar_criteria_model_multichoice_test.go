package bodycodegen

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
)

const multiChoiceScalarSource = `package multiroute
namespace multiroute
entity Integer id "urn:multiroute:integer"
entity Boolean id "urn:multiroute:boolean"
activity Double(Integer) -> Integer id "urn:multiroute:double" computes "let first=input; let second=input+3; first=first*2; first=first+5; return second" assembling {
 choice "z_target" assignment_target at "0" alternative "second" intent "Assign the multiplication result to second."
 choice "a_schedule" root_order at "2" intent "Keep multiplication before incrementing first."
 criteria { predicate "double" activity "IsDouble" input "[-6]" input "[7]" holdout_input "[15]" }
 attempts "2"
}
activity IsDouble(Integer,Integer) -> Boolean id "urn:multiroute:predicate" computes "return input1 == input0*2"
`

func TestScalarSourceModelCombinesAssignmentsAndOrderWithoutSortingChoices(t *testing.T) {
	g, err := NewTypedPathGenerator(scalarBridgeSyntheticArtifact(t))
	if err != nil {
		t.Fatal(err)
	}
	source := []byte(multiChoiceScalarSource)
	result, err := g.GenerateSourceAssembly(context.Background(), "multiroute.gooo", source, "Double")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r == nil || r.ScalarPrediction == nil || r.ModelCalls != 1 || r.ScalarPrediction.ChoiceCount != 2 ||
		r.ScalarPrediction.CandidateCount != 4 || r.ScalarPrediction.ProposedMask != 3 || r.SelectedMask != 1 ||
		!slices.Equal(r.Ranking, []uint16{3, 1, 2, 0}) || len(r.Attempts) != 2 || r.Status != "COMPLETE_FINITE" {
		t.Fatal("joint source ranking or finite selection changed", r)
	}
	if r.Choices[0].Kind != "assignment_target" || r.Choices[1].Kind != "root_order" ||
		r.Attempts[0].Mask != 3 || r.Attempts[0].Criteria.Passed != 0 || r.Attempts[0].Criteria.Total != 2 ||
		r.Attempts[1].Mask != 1 || r.Criteria.Passed != 2 || r.Criteria.Total != 2 ||
		r.CriteriaHoldout.Passed != 1 || r.CriteriaHoldout.Total != 1 || result.Report.ActivityID != "urn:multiroute:double" {
		t.Fatal("choice declaration order, first failure or actual conditions changed", r)
	}
	logMultiChoiceScalarObservation(t, r)
}

func logMultiChoiceScalarObservation(t *testing.T, r *RecordAssemblyReceipt) {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"scope": "one fresh assignment/order source with intentionally reversed lexical choice names",
		"source_sha256": r.OriginalSourceSHA256, "choices": len(r.Choices), "candidate_count": r.ScalarPrediction.CandidateCount,
		"proposed_mask": r.ScalarPrediction.ProposedMask, "selected_mask": r.SelectedMask, "ranking": r.Ranking,
		"first_proposal_passed": r.Attempts[0].Criteria.Passed, "first_proposal_total": r.Attempts[0].Criteria.Total,
		"selection_passed": r.Criteria.Passed, "selection_total": r.Criteria.Total,
		"held_passed": r.CriteriaHoldout.Passed, "held_total": r.CriteriaHoldout.Total, "model_calls": r.ModelCalls})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("MULTICHOICE_SCALAR_OBSERVATION", string(raw))
}
