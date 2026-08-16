package pressureindependence

import (
	"testing"

	workfrontier "github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

func TestCurrentV1SameGroupFalseAcceptance(t *testing.T) {
	input := mustCorpusInput(t, "two-ids-one-group-unknown")
	oracle := Evaluate(input)
	current := workfrontier.Select(currentV1Input(input))
	if oracle.Decision != DecisionUnknown || oracle.Reason != ReasonIndependentGroupShortfall {
		t.Fatalf("independent oracle = %#v", oracle)
	}
	if current.Status != workfrontier.DecisionPass || len(current.SelectedIDs) != 1 {
		t.Fatalf("current V1 selector = %#v", current)
	}
	t.Logf("current-V1 false acceptance status=%s selected=%v; independent decision=%s reason=%s",
		current.Status, current.SelectedIDs, oracle.Decision, oracle.Reason)
}

func currentV1Input(input Input) workfrontier.Input {
	pressures := make([]workfrontier.Pressure, 0, len(input.RequiredPressureIDs))
	for _, id := range input.RequiredPressureIDs {
		pressures = append(pressures, workfrontier.Pressure{StableID: id})
	}
	return workfrontier.Input{
		SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: input.AuthoritySnapshotDigest,
		PolicyDigest: input.PolicyDigest, RegistryDigest: input.RegistryDigest,
		MinimumSelectedPressures: 2, Capacity: workfrontier.Capacity{CPUCoreNS: 10},
		Pressures: pressures,
		States:    []workfrontier.ObligationState{{ObligationID: "obligation-a", Status: "PENDING"}},
		Paths: []workfrontier.RepairPath{{
			StableID: "path-a", WorkID: "work-a", ObligationID: "obligation-a",
			ReadSet: []string{input.RequiredPressureIDs[0]}, WriteSet: []string{input.RequiredPressureIDs[1]},
			RequiredPressureIDs: append([]string(nil), input.RequiredPressureIDs...), CPUCoreNSUpperBound: 1,
		}},
	}
}
