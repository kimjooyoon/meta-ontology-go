package workfrontier

import (
	"encoding/json"
	"sort"
	"testing"
)

func independentOracle(t *testing.T, raw []byte) oracleResult {
	t.Helper()
	var input oracleInput
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatal(err)
	}
	if input.Pressures == nil || input.M < 2 || input.M != len(*input.Pressures) {
		return oracleResult{Status: "UNKNOWN"}
	}
	registered := make(map[string]bool, len(input.RegisteredPaths))
	for _, path := range input.RegisteredPaths {
		registered[path] = true
	}
	ready := make([]oraclePressure, 0, len(*input.Pressures))
	blocked, unknown := false, false
	for _, pressure := range *input.Pressures {
		if pressure.ID == "" || pressure.WorkID == "" || pressure.CPU <= 0 {
			unknown = true
			continue
		}
		if pressure.Prerequisite != "PASS" || pressure.CPU > input.CPUCapacity {
			blocked = true
			continue
		}
		if !validClaims(pressure.Claims, registered) {
			unknown = true
			continue
		}
		ready = append(ready, pressure)
	}
	if unknown {
		return oracleResult{Status: "UNKNOWN"}
	}
	sort.Slice(ready, func(i, j int) bool {
		if ready[i].Priority != ready[j].Priority {
			return ready[i].Priority < ready[j].Priority
		}
		return ready[i].ID < ready[j].ID
	})
	selected := make([]oraclePressure, 0, len(ready))
	usedCPU := 0
	for _, pressure := range ready {
		if usedCPU+pressure.CPU > input.CPUCapacity || conflictsWithAny(pressure, selected) {
			continue
		}
		selected = append(selected, pressure)
		usedCPU += pressure.CPU
	}
	if len(selected) == 0 && blocked {
		return oracleResult{Status: "BLOCKED"}
	}
	result := oracleResult{Status: "PASS", MaximumSize: maximumCompatibleSize(ready, input.CPUCapacity)}
	for _, pressure := range selected {
		result.SelectedIDs = append(result.SelectedIDs, pressure.ID)
		result.WorkIDs = append(result.WorkIDs, pressure.WorkID)
	}
	return result
}
