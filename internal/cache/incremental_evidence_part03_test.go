package cache

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strconv"
	"testing"
)

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
