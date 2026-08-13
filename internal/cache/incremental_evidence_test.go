package cache

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	incrementalPartCount = 10
)

var incrementalFixtureSizes = []int{10, 100, 1000, 10000}

type incrementalFixture struct {
	size  int
	ir    semantic.IR
	facts []semantic.Fact
}

type incrementalPart struct {
	index         int
	closureDigest Digest
	spec          PartialSpec
	key           Key
}

type incrementalMutation struct {
	name             string
	presentationOnly bool
	mutatedFactIndex int
	dependencyPart   int
	expectedAffected func(size int) map[int]bool
}

type incrementalMeasurement struct {
	hits           int
	misses         int
	recomputations int
	started        time.Time
}

func TestIncrementalCacheMutationMatrix(t *testing.T) {
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

func BenchmarkIncrementalSemanticDigest(b *testing.B) {
	for _, size := range incrementalFixtureSizes {
		size := size
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			fixture := newIncrementalFixture(b, size, false, -1)
			b.ReportAllocs()
			b.ReportMetric(float64(size), "facts/op")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := SemanticDigest(fixture.ir); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func newIncrementalFixture(t testing.TB, size int, presentationOnly bool, mutatedFactIndex int) incrementalFixture {
	t.Helper()
	fixture := incrementalFixture{size: size, ir: semantic.NewIR("cache-incremental-fixture", semantic.Namespace("cache-perf")), facts: make([]semantic.Fact, size)}
	for index := 0; index < size; index++ {
		activityID := semantic.MustIdentity(fmt.Sprintf("fixture://activity/%05d", index))
		entityID := semantic.MustIdentity(fmt.Sprintf("fixture://entity/%05d", index))
		if index == mutatedFactIndex {
			entityID = semantic.MustIdentity("fixture://entity/mutated")
		}
		activityName := fmt.Sprintf("Activity %05d", index)
		entityName := fmt.Sprintf("Entity %05d", index)
		if presentationOnly {
			activityName = fmt.Sprintf("Renamed activity %05d", index)
			entityName = fmt.Sprintf("Renamed entity %05d", index)
		}
		activity, err := semantic.NewActivity(activityID, semantic.Namespace("cache-perf"), activityName)
		if err != nil {
			t.Fatal(err)
		}
		entity, err := semantic.NewEntity(entityID, semantic.Namespace("cache-perf"), entityName)
		if err != nil {
			t.Fatal(err)
		}
		if err := fixture.ir.AddNode(activity); err != nil {
			t.Fatal(err)
		}
		if err := fixture.ir.AddNode(entity); err != nil {
			t.Fatal(err)
		}
		fact := semantic.NewUsedFact(activityID, entityID)
		if err := fixture.ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
		fixture.facts[index] = fact
	}
	if err := fixture.ir.Validate(); err != nil {
		t.Fatalf("%d-fact fixture invalid: %v", size, err)
	}
	return fixture
}

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
		parts[partIndex] = incrementalPart{index: partIndex, closureDigest: closureDigest, spec: spec, key: key}
	}
	return parts
}

func assertIncrementalOracle(t *testing.T, base, current []incrementalPart, expected map[int]bool, mutation string) {
	t.Helper()
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

func measureIncrementalCache(t *testing.T, base, current []incrementalPart, expected map[int]bool, size int, mutation string) incrementalMeasurement {
	t.Helper()
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range base {
		if err := cache.PutPartial(part.spec, []byte(fmt.Sprintf("base:%d:%d", size, part.index))); err != nil {
			t.Fatal(err)
		}
	}
	measurement := incrementalMeasurement{started: time.Now()}
	for _, part := range current {
		part := part
		_, data, _, hit, err := cache.GetOrComputePartial(context.Background(), part.spec, func() ([]byte, error) {
			measurement.recomputations++
			return []byte(fmt.Sprintf("current:%d:%d:%s", size, part.index, mutation)), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		wantHit := !expected[part.index]
		if hit != wantHit {
			t.Fatalf("%s part %d hit=%v, want %v", mutation, part.index, hit, wantHit)
		}
		if hit {
			measurement.hits++
			if string(data) != fmt.Sprintf("base:%d:%d", size, part.index) {
				t.Fatalf("%s part %d hit returned %q", mutation, part.index, data)
			}
		} else {
			measurement.misses++
			if string(data) != fmt.Sprintf("current:%d:%d:%s", size, part.index, mutation) {
				t.Fatalf("%s part %d miss returned %q", mutation, part.index, data)
			}
		}
	}
	if measurement.misses != len(expected) || measurement.recomputations != measurement.misses {
		t.Fatalf("%s measurements hits=%d misses=%d recomputations=%d expected misses=%d", mutation, measurement.hits, measurement.misses, measurement.recomputations, len(expected))
	}
	return measurement
}

func singleAffectedPart(size int) map[int]bool {
	return map[int]bool{(size / 2) * incrementalPartCount / size: true}
}

func uniqueSortedStrings(values []string) []string {
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	unique := sorted[:0]
	for _, value := range sorted {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}
