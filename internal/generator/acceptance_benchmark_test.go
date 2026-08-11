package generator

import "testing"

func BenchmarkAcceptanceFixtureGeneration(b *testing.B) {
	ir := acceptanceFixture()
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		if _, err := Generate(ir, nil); err != nil {
			b.Fatal(err)
		}
	}
}
