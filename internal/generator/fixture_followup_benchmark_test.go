package generator

import "testing"

func BenchmarkFixtureRemovalLocality(b *testing.B) {
	first, err := Generate(acceptanceFixture(), nil)
	if err != nil {
		b.Fatal(err)
	}
	changed := acceptanceFixture()
	changed.Activities = changed.Activities[:1]
	b.ResetTimer()
	for range b.N {
		if _, err := Generate(changed, first.Source); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFixtureDeclarationPermutation(b *testing.B) {
	ir := acceptanceFixture()
	ir.Entities[0], ir.Entities[1] = ir.Entities[1], ir.Entities[0]
	ir.Activities[0], ir.Activities[1] = ir.Activities[1], ir.Activities[0]
	b.ResetTimer()
	for range b.N {
		if _, err := Generate(ir, nil); err != nil {
			b.Fatal(err)
		}
	}
}
