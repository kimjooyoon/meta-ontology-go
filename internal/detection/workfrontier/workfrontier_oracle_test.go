package workfrontier

import (
	"encoding/json"
	"sort"
	"testing"
)

type contractFixture struct {
	Name              string          `json:"name"`
	Input             json.RawMessage `json:"input"`
	DecodeError       bool            `json:"decode_error"`
	PermutationTest   bool            `json:"permutation_test"`
	FairBaseline      bool            `json:"fair_baseline"`
	RequiredConflicts [][]string      `json:"required_conflicts"`
	Expected          expectedResult  `json:"expected"`
}

type expectedResult struct {
	Status      string   `json:"status"`
	Quality     string   `json:"quality"`
	SelectedIDs []string `json:"selected_ids"`
	WorkIDs     []string `json:"work_ids"`
}

type oracleInput struct {
	M               int               `json:"m"`
	CPUCapacity     int               `json:"cpu_capacity"`
	RegisteredPaths []string          `json:"registered_paths"`
	Pressures       *[]oraclePressure `json:"pressures"`
}

type oraclePressure struct {
	ID           string        `json:"id"`
	WorkID       string        `json:"work_id"`
	Priority     int           `json:"priority"`
	CPU          int           `json:"cpu"`
	Prerequisite string        `json:"prerequisite"`
	Claims       []oracleClaim `json:"claims"`
}

type oracleClaim struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
}

type oracleResult struct {
	Status      string
	SelectedIDs []string
	WorkIDs     []string
}

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
	result := oracleResult{Status: "PASS"}
	for _, pressure := range selected {
		result.SelectedIDs = append(result.SelectedIDs, pressure.ID)
		result.WorkIDs = append(result.WorkIDs, pressure.WorkID)
	}
	return result
}

func validClaims(claims []oracleClaim, registered map[string]bool) bool {
	if len(claims) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, claim := range claims {
		if claim.Path == "" || !registered[claim.Path] || (claim.Mode != "R" && claim.Mode != "W") || seen[claim.Path] {
			return false
		}
		seen[claim.Path] = true
	}
	return true
}

func conflictsWithAny(candidate oraclePressure, selected []oraclePressure) bool {
	for _, prior := range selected {
		if pressuresConflict(candidate, prior) {
			return true
		}
	}
	return false
}

func pressuresConflict(left, right oraclePressure) bool {
	for _, leftClaim := range left.Claims {
		for _, rightClaim := range right.Claims {
			if leftClaim.Path == rightClaim.Path && (leftClaim.Mode == "W" || rightClaim.Mode == "W") {
				return true
			}
		}
	}
	return false
}

func assertRequiredConflicts(t *testing.T, fixture contractFixture) {
	t.Helper()
	if len(fixture.RequiredConflicts) == 0 {
		return
	}
	var input oracleInput
	if err := json.Unmarshal(fixture.Input, &input); err != nil || input.Pressures == nil {
		t.Fatalf("conflict fixture input = %v", err)
	}
	byID := make(map[string]oraclePressure, len(*input.Pressures))
	for _, pressure := range *input.Pressures {
		byID[pressure.ID] = pressure
	}
	for _, pair := range fixture.RequiredConflicts {
		if len(pair) != 2 || !pressuresConflict(byID[pair[0]], byID[pair[1]]) {
			t.Fatalf("oracle missed required conflict %v", pair)
		}
	}
}
