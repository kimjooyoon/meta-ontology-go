package workfrontier

import (
	"encoding/json"
	"testing"
)

func legacyClaims(claims []oracleClaim, registered map[string]struct{}) ([]string, []string, bool) {
	readSet := make([]string, 0, len(claims))
	writeSet := make([]string, 0, len(claims))
	seen := make(map[string]struct{}, len(claims))
	valid := len(claims) != 0
	for _, claim := range claims {
		if claim.Path == "" || claim.Mode != "R" && claim.Mode != "W" {
			valid = false
			continue
		}
		if _, duplicate := seen[claim.Path]; duplicate {
			valid = false
		}
		seen[claim.Path] = struct{}{}
		if _, ok := registered[claim.Path]; !ok {
			valid = false
		}
		if claim.Mode == "R" {
			readSet = append(readSet, claim.Path)
		} else {
			writeSet = append(writeSet, claim.Path)
		}
	}
	return readSet, writeSet, valid
}
func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
func loadFixtures(t *testing.T) []contractFixture {
	t.Helper()
	data, err := contractFixtures.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []contractFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}
