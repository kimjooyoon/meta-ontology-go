package cache

import "testing"

// assertSemanticDeltaOracle compares semantic facts and affected closure
// independently from cache-key equality, preventing a broken key from making
// its own oracle pass.
func assertSemanticDeltaOracle(t *testing.T, base, current []incrementalPart, expected map[int]bool, mutation string) {
	t.Helper()
	changedFacts := make(map[int]bool)
	for index := range base {
		if !equalStrings(base[index].factDigests, current[index].factDigests) {
			changedFacts[index] = true
		}
	}
	if mutation == "local fact" {
		if len(changedFacts) != 1 {
			t.Fatalf("direct semantic delta changed parts=%v, want one part", changedFacts)
		}
	} else if len(changedFacts) != 0 {
		t.Fatalf("%s direct semantic delta changed parts=%v, want none", mutation, changedFacts)
	}
	for index := range base {
		affected := !equalStrings(base[index].closureIDs, current[index].closureIDs)
		if affected != expected[index] && mutation != "dependency closure" {
			t.Fatalf("%s direct affected closure part %d=%v, want %v", mutation, index, affected, expected[index])
		}
		if mutation == "dependency closure" && affected {
			t.Fatalf("dependency-only mutation changed semantic closure for part %d", index)
		}
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
