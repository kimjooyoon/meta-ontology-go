package fuzz

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var benchmarkDiagnosticCount int

func BenchmarkContractFixtures(b *testing.B) {
	manifest := loadContractManifest(b)
	sources := make([]string, len(manifest.Fixtures))
	for index, fixture := range manifest.Fixtures {
		source, err := os.ReadFile("testdata/" + fixture.File)
		if err != nil {
			b.Fatal(err)
		}
		sources[index] = string(source)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, source := range sources {
			_, diagnostics := syntax.ParseFile(fuzzFilename, source)
			benchmarkDiagnosticCount += len(diagnostics)
		}
	}
}
