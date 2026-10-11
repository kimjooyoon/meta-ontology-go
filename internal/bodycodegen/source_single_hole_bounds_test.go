package bodycodegen

import (
	"fmt"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestSourceSingleHoleFillBounds(t *testing.T) {
	for _, count := range []int{0, 1, 8, 9} {
		plan := IRBodyFillPlan{Schema: bodyFillMultiPlanSchema, Intent: "Retain each declared assignment.",
			TestCases: []IRBodyFillTestCase{{Input: 1, Expected: 2}}}
		spec := assemblyspec.Spec{Cases: []assemblyspec.Case{{Input: 1, Expected: 2}},
			FillPlan: &assemblyspec.FillPlan{Intent: plan.Intent}}
		for i := range count {
			id := fmt.Sprintf("hole%d", i)
			plan.Holes = append(plan.Holes, IRBodyFillHole{ID: id})
			spec.FillPlan.Holes = append(spec.FillPlan.Holes, assemblyspec.FillHole{ID: id})
		}
		for _, id := range []string{"first", "second"} {
			candidate := IRBodyFillCandidate{ID: id, Fills: map[string]string{}}
			owned := assemblyspec.FillCandidate{ID: id}
			for _, hole := range plan.Holes {
				candidate.Fills[hole.ID] = "input + 1"
				owned.Fills = append(owned.Fills, assemblyspec.HoleFilling{HoleID: hole.ID, Expression: "input + 1"})
			}
			plan.Candidates = append(plan.Candidates, candidate)
			spec.FillPlan.Candidates = append(spec.FillPlan.Candidates, owned)
		}
		want := count >= 1 && count <= 8
		if err := validateIRBodyFillPlan(plan); (err == nil) != want {
			t.Fatalf("IR hole count %d: %v", count, err)
		}
		if err := spec.Validate(); (err == nil) != want {
			t.Fatalf("integer source hole count %d: %v", count, err)
		}
		spec.Cases = nil
		spec.ValueCases = []assemblyspec.ValueCase{{Inputs: `[{"text":"a"}]`, Expected: `{"offset":1}`}}
		if err := spec.Validate(); (err == nil) != want {
			t.Fatalf("record source hole count %d: %v", count, err)
		}
	}
}
