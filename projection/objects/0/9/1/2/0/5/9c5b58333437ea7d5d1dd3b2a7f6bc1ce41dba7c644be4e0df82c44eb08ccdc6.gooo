package cache

import (
	"fmt"
	"testing"
)

func newIncrementalParts(t testing.TB, fixture incrementalFixture, dependencyMutationPart int) []incrementalPart {
	t.Helper()
	parts := make([]incrementalPart, incrementalPartCount)
	for partIndex := range parts {
		start := partIndex * fixture.size / incrementalPartCount
		end := (partIndex + 1) * fixture.size / incrementalPartCount
		factDigests := make([]string, 0, end-start)
		closureIDs := make([]string, 0, 2*(end-start))
		for _, fact := range fixture.facts[start:end] {
			factDigests = append(factDigests, fact.StableHash())
			closureIDs = append(closureIDs, fact.Subject.String(), fact.Object.String())
		}
		closureIDs = uniqueSortedStrings(closureIDs)
		closureDigest, err := DigestOf(closureIDs)
		if err != nil {
			t.Fatal(err)
		}
		dependencies := map[string]any{"facts": factDigests, "closure": closureIDs}
		if partIndex == dependencyMutationPart {
			dependencies["dependency_mutation"] = "changed"
		}
		spec := PartialSpec{Part: fmt.Sprintf("part-%02d", partIndex), KeySpec: KeySpec{
			Version: "v1", Namespace: "cache-perf", ToolVersion: "compiler-1",
			Inputs: struct {
				Facts   []string
				Closure []string
			}{factDigests, closureIDs},
			OptionsDigest: mustOptionsDigest(map[string]any{"mode": "incremental-evidence"}), Freshness: FreshnessSpec{
				Dependencies: dependencies, Provenance: map[string]any{"fixture": fixture.size, "part": partIndex},
			},
		}}
		key, err := NewPartialKey(spec)
		if err != nil {
			t.Fatal(err)
		}
		parts[partIndex] = incrementalPart{index: partIndex, factDigests: factDigests, closureIDs: closureIDs,
			closureDigest: closureDigest, spec: spec, key: key}
	}
	return parts
}
func assertIncrementalOracle(t *testing.T, base, current []incrementalPart, expected map[int]bool, mutation string) {
	t.Helper()
	assertSemanticDeltaOracle(t, base, current, expected, mutation)
	for index := range base {
		keyChanged := base[index].key != current[index].key
		if keyChanged != expected[index] {
			t.Fatalf("%s part %d keyChanged=%v, want %v", mutation, index, keyChanged, expected[index])
		}
		closureChanged := base[index].closureDigest != current[index].closureDigest
		if mutation == "local fact" && closureChanged != expected[index] {
			t.Fatalf("%s part %d closureChanged=%v, want %v", mutation, index, closureChanged, expected[index])
		}
		if mutation != "local fact" && closureChanged {
			t.Fatalf("%s changed semantic closure for part %d", mutation, index)
		}
	}
}
