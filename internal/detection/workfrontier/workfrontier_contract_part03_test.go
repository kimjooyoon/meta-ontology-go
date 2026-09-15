package workfrontier

import (
	"encoding/json"
	"fmt"
	"testing"
)

func adaptLegacyInput(t *testing.T, raw []byte) []byte {
	t.Helper()
	var legacy oracleInput
	if err := json.Unmarshal(raw, &legacy); err != nil {
		t.Fatal(err)
	}
	input := map[string]any{
		"schema_version":             SchemaVersion,
		"snapshot_digest":            "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"policy_digest":              "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"registry_digest":            "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		"minimum_selected_pressures": legacy.M,
		"capacity":                   map[string]any{"cpu_core_ns": legacy.CPUCapacity},
		"pressures":                  nil,
		"states":                     nil,
		"paths":                      nil,
	}
	if legacy.Pressures == nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	registered := make(map[string]struct{}, len(legacy.RegisteredPaths))
	pressures := make([]Pressure, 0, len(legacy.RegisteredPaths)+maxInt(legacy.M, 0))
	for _, path := range legacy.RegisteredPaths {
		registered[path] = struct{}{}
		pressures = append(pressures, Pressure{StableID: path})
	}
	requiredPressureIDs := make([]string, 0, maxInt(legacy.M, 0))
	for index := 0; index < legacy.M; index++ {
		id := fmt.Sprintf("pressure/registry/%d", index)
		requiredPressureIDs = append(requiredPressureIDs, id)
		pressures = append(pressures, Pressure{StableID: id})
	}
	states := make([]ObligationState, 0, len(*legacy.Pressures)*2)
	paths := make([]RepairPath, 0, len(*legacy.Pressures))
	for _, candidate := range *legacy.Pressures {
		obligationID := "obligation/" + candidate.ID
		prerequisiteID := "prerequisite/" + candidate.ID
		states = append(states,
			ObligationState{ObligationID: obligationID, Status: "PENDING"},
			ObligationState{ObligationID: prerequisiteID, Status: candidate.Prerequisite},
		)
		readSet, writeSet, claimsValid := legacyClaims(candidate.Claims, registered)
		required := append([]string(nil), requiredPressureIDs...)
		if !claimsValid || candidate.ID == "" || candidate.WorkID == "" || candidate.CPU <= 0 {
			required = append(required, "pressure/unresolved/"+candidate.ID)
		}
		paths = append(paths, RepairPath{
			StableID: candidate.ID, WorkID: candidate.WorkID, ObligationID: obligationID,
			PrerequisiteObligationIDs: []string{prerequisiteID}, ReadSet: readSet,
			WriteSet: writeSet, RequiredPressureIDs: required,
			PolicyPriority: uint32(maxInt(candidate.Priority, 0)), CPUCoreNSUpperBound: uint64(maxInt(candidate.CPU, 0)),
		})
	}
	input["pressures"] = pressures
	input["states"] = states
	input["paths"] = paths
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
