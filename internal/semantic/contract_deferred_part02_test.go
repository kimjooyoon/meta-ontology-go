package semantic

import (
	"errors"
	"testing"
)

func TestActivityContractRejectsInvalidFactsAtomically(t *testing.T) {
	ns := Namespace("contract-boundary")
	activity := mustActivity(t, MustIdentity("contract-boundary://activity/run"), ns, "Run")
	input := mustEntity(t, MustIdentity("contract-boundary://entity/input"), ns, "Input")
	output := mustEntity(t, MustIdentity("contract-boundary://entity/output"), ns, "Output")
	otherActivity := mustActivity(t, MustIdentity("contract-boundary://activity/other"), ns, "Other")
	agent := mustAgent(t, MustIdentity("contract-boundary://agent/verifier"), ns, "Verifier")
	missing := MustIdentity("contract-boundary://entity/missing")
	for _, test := range []struct {
		name     string
		contract ActivityContract
		wantErr  error
	}{
		{
			name: "missing input endpoint", contract: ActivityContract{
				Activity: activity.ID, Inputs: []ID{input.ID, missing},
			}, wantErr: ErrNodeNotFound,
		},
		{
			name: "missing output endpoint", contract: ActivityContract{
				Activity: activity.ID, Outputs: []ID{output.ID, missing},
			}, wantErr: ErrNodeNotFound,
		},
		{
			name: "wrong input kind", contract: ActivityContract{
				Activity: activity.ID, Inputs: []ID{input.ID, otherActivity.ID},
			}, wantErr: ErrInvalidFact,
		},
		{
			name: "wrong output kind", contract: ActivityContract{
				Activity: activity.ID, Outputs: []ID{output.ID, otherActivity.ID},
			}, wantErr: ErrInvalidFact,
		},
		{
			name: "missing agent endpoint", contract: ActivityContract{
				Activity: activity.ID, Agents: []ID{agent.ID, missing},
			}, wantErr: ErrNodeNotFound,
		},
		{
			name: "wrong agent kind", contract: ActivityContract{
				Activity: activity.ID, Agents: []ID{agent.ID, input.ID},
			}, wantErr: ErrInvalidFact,
		},
	} {
		ir := NewIR("contract-boundary", ns)
		for _, node := range []Node{activity, input, output, otherActivity, agent} {
			if err := ir.AddNode(node); err != nil {
				t.Fatal(err)
			}
		}
		before := ir.Canonical()
		if err := ir.AddActivityContract(test.contract); !errors.Is(err, test.wantErr) {
			t.Fatalf("%s error = %v, want %v", test.name, err, test.wantErr)
		}
		if ir.Canonical() != before || len(ir.Graph.Facts()) != 0 || len(ir.Evidence()) != 0 {
			t.Fatalf("%s partially mutated graph, IR, or evidence", test.name)
		}
	}
}
