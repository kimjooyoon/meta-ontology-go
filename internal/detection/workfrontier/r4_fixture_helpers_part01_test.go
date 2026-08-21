package workfrontier

import (
	"testing"
)

func r4FixtureInput(t *testing.T, name string) R4Input {
	t.Helper()
	base := R4Input{
		SchemaVersion:            R4SchemaVersion,
		MinimumSelectedPressures: 2,
		Capacity:                 Capacity{CPUCoreNS: 20},
		Pressures:                []Pressure{{StableID: "pressure/a"}, {StableID: "pressure/b"}},
		States:                   []ObligationState{{ObligationID: "obligation/root", Status: "PENDING"}},
		Paths: []RepairPath{{
			StableID: "path/root", ObligationID: "obligation/root", ReadSet: []string{"pressure/a"},
			WriteSet: []string{"pressure/a"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"},
			PolicyPriority: 1, CPUCoreNSUpperBound: 1,
		}},
		RootObligationIDs: []string{"obligation/root"},
		Rules:             []R4Rule{},
	}
	switch name {
	case "acyclic":
		base.States = append(base.States, ObligationState{ObligationID: "obligation/child", Status: "PENDING"})
		base.Paths = append(base.Paths, RepairPath{
			StableID: "path/child", ObligationID: "obligation/child", PrerequisiteObligationIDs: []string{"obligation/root"},
			ReadSet: []string{"pressure/b"}, WriteSet: []string{"pressure/b"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"},
			PolicyPriority: 2, CPUCoreNSUpperBound: 1,
		})
	case "self-loop", "missing-bound", "zero-bound", "stale-digest", "iteration-exhaustion", "conflicting-rule":
		base.Paths[0].PrerequisiteObligationIDs = []string{"obligation/root"}
		base = bindR4Rules(t, base, name)
	case "two-node-cycle", "reordered-input":
		base.States = append(base.States, ObligationState{ObligationID: "obligation/b", Status: "PENDING"})
		base.Paths[0] = RepairPath{
			StableID: "path/a", ObligationID: "obligation/root", PrerequisiteObligationIDs: []string{"obligation/b"},
			ReadSet: []string{"pressure/a"}, WriteSet: []string{"pressure/a"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"},
			PolicyPriority: 1, CPUCoreNSUpperBound: 1,
		}
		base.Paths = append(base.Paths, RepairPath{
			StableID: "path/b", ObligationID: "obligation/b", PrerequisiteObligationIDs: []string{"obligation/root"},
			ReadSet: []string{"pressure/b"}, WriteSet: []string{"pressure/b"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"},
			PolicyPriority: 2, CPUCoreNSUpperBound: 1,
		})
		base = bindR4Rules(t, base, name)
	case "unreachable-cycle":
		base.States = append(base.States,
			ObligationState{ObligationID: "obligation/a", Status: "PENDING"},
			ObligationState{ObligationID: "obligation/b", Status: "PENDING"},
		)
		base.Paths = append(base.Paths,
			RepairPath{StableID: "path/a", ObligationID: "obligation/a", PrerequisiteObligationIDs: []string{"obligation/b"}, ReadSet: []string{"pressure/a"}, WriteSet: []string{"pressure/a"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"}, CPUCoreNSUpperBound: 1},
			RepairPath{StableID: "path/b", ObligationID: "obligation/b", PrerequisiteObligationIDs: []string{"obligation/a"}, ReadSet: []string{"pressure/b"}, WriteSet: []string{"pressure/b"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"}, CPUCoreNSUpperBound: 1},
		)
	default:
		t.Fatalf("unknown R4 fixture %q", name)
	}
	bound, err := BindR4Payloads(base)
	if err != nil {
		t.Fatal(err)
	}
	return bound
}
