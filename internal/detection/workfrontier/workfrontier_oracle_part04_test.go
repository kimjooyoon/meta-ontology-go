package workfrontier

import (
	"encoding/json"
	"testing"
)

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
