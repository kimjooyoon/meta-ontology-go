package generator

import "testing"

func BenchmarkFixtureMultipleSlotPreservation(b *testing.B) {
	first, err := Generate(acceptanceFixture(), nil)
	if err != nil {
		b.Fatal(err)
	}
	previous := []byte(first.Source)
	changed := acceptanceFixture()
	changed.Activities[0].Slots[0].Default = "panic(\"new compile default\")"
	changed.Activities[1].Slots[0].Default = "panic(\"new inspect default\")"
	b.ResetTimer()
	for range b.N {
		if _, err := Generate(changed, previous); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFixtureSourceMapBounds(b *testing.B) {
	ir := acceptanceFixture()
	b.ResetTimer()
	for range b.N {
		result, err := Generate(ir, nil)
		if err != nil || len(result.SourceMap.Mappings) == 0 {
			b.Fatalf("source-map generation failed: %v", err)
		}
	}
}
