package workfrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

const (
	r4FixtureSnapshotPayload = `{"fixture":"r4","kind":"snapshot","revision":1}`
	r4FixturePolicyPayload   = `{"fixture":"r4","kind":"policy","revision":1}`
	r4FixtureRegistryPayload = `{"fixture":"r4","kind":"registry","revision":1}`
)

func r4FixtureInput(t *testing.T, name string) R4Input {
	t.Helper()
	base := R4Input{
		SchemaVersion:            R4SchemaVersion,
		SnapshotDigest:           r4BindingDigest(r4FixtureSnapshotPayload),
		SnapshotPayload:          r4FixtureSnapshotPayload,
		PolicyDigest:             r4BindingDigest(r4FixturePolicyPayload),
		PolicyPayload:            r4FixturePolicyPayload,
		RegistryDigest:           r4BindingDigest(r4FixtureRegistryPayload),
		RegistryPayload:          r4FixtureRegistryPayload,
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
	return base
}

func bindR4Rules(t *testing.T, input R4Input, name string) R4Input {
	t.Helper()
	if name == "missing-bound" {
		return input
	}
	graph, err := AnalyzeR4Graph(input)
	if err != nil {
		t.Fatal(err)
	}
	var digest string
	for _, component := range graph.SCCs {
		if component.Cyclic {
			digest = component.Digest
		}
	}
	if name == "stale-digest" {
		digest = "stale-scc-digest"
	}
	maxIterations := uint64(2)
	iterationsUsed := uint64(0)
	if name == "zero-bound" {
		maxIterations = 0
	}
	if name == "iteration-exhaustion" {
		maxIterations = 1
		iterationsUsed = 1
	}
	input.Rules = []R4Rule{{SCCDigest: digest, MaxIterations: maxIterations, IterationsUsed: iterationsUsed}}
	if name == "conflicting-rule" {
		input.Rules = append(input.Rules, R4Rule{SCCDigest: digest, MaxIterations: maxIterations + 1})
	}
	return input
}

func r4BindingDigest(payload string) string {
	digest := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(digest[:])
}

func r4FixtureDigest(input R4Input) string {
	data, err := EncodeR4JSON(input)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func reversePressures(values []Pressure) []Pressure {
	result := append([]Pressure(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseStates(values []ObligationState) []ObligationState {
	result := append([]ObligationState(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reversePaths(values []RepairPath) []RepairPath {
	result := append([]RepairPath(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseRules(values []R4Rule) []R4Rule {
	result := append([]R4Rule(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
