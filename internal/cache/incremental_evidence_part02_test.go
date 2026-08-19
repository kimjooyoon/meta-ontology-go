package cache

import (
	"strconv"
	"testing"
	"time"
)

func TestIncrementalCacheMutationMatrix(t *testing.T) {
	requireClassifiedCacheTest(t, "TestIncrementalCacheMutationMatrix", CacheTestClassSlowObservation)
	mutations := []incrementalMutation{
		{name: "presentation rename", presentationOnly: true, mutatedFactIndex: -1, dependencyPart: -1, expectedAffected: func(int) map[int]bool { return map[int]bool{} }},
		{name: "local fact", mutatedFactIndex: -1, dependencyPart: -1, expectedAffected: singleAffectedPart},
		{name: "dependency closure", mutatedFactIndex: -1, dependencyPart: 3, expectedAffected: func(int) map[int]bool { return map[int]bool{3: true} }},
	}
	for _, size := range incrementalFixtureSizes {
		size := size
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			base := newIncrementalFixture(t, size, false, -1)
			baseDigest, err := SemanticDigest(base.ir)
			if err != nil {
				t.Fatal(err)
			}
			repeated := newIncrementalFixture(t, size, false, -1)
			repeatedDigest, err := SemanticDigest(repeated.ir)
			if err != nil {
				t.Fatal(err)
			}
			if repeatedDigest != baseDigest {
				t.Fatalf("non-deterministic %d-fact fixture: %s != %s", size, baseDigest, repeatedDigest)
			}
			baseParts := newIncrementalParts(t, base, -1)
			for _, mutation := range mutations {
				mutation := mutation
				t.Run(mutation.name, func(t *testing.T) {
					mutatedIndex := mutation.mutatedFactIndex
					if mutatedIndex < 0 && mutation.name == "local fact" {
						mutatedIndex = size / 2
					}
					current := newIncrementalFixture(t, size, mutation.presentationOnly, mutatedIndex)
					currentDigest, err := SemanticDigest(current.ir)
					if err != nil {
						t.Fatal(err)
					}
					if mutation.presentationOnly && currentDigest != baseDigest {
						t.Fatalf("presentation mutation changed semantic digest: %s != %s", currentDigest, baseDigest)
					}
					if !mutation.presentationOnly && mutation.name == "local fact" && currentDigest == baseDigest {
						t.Fatal("local fact mutation retained semantic digest")
					}
					currentParts := newIncrementalParts(t, current, mutation.dependencyPart)
					expected := mutation.expectedAffected(size)
					assertIncrementalOracle(t, baseParts, currentParts, expected, mutation.name)
					measure := measureIncrementalCache(t, baseParts, currentParts, expected, size, mutation.name)
					t.Logf("facts=%d mutation=%q hits=%d misses=%d recomputations=%d elapsed=%s", size, mutation.name, measure.hits, measure.misses, measure.recomputations, time.Since(measure.started))
				})
			}
		})
	}
}
