package generator

import "testing"

func BenchmarkFixtureGeneratedRegionLocality(b *testing.B) {
	first, err := Generate(acceptanceFixture(), nil)
	if err != nil {
		b.Fatal(err)
	}
	changed := acceptanceFixture()
	changed.Activities[0].GoName = "CompileBootstrap"
	changed.Activities[0].Name = "CompileBootstrap"
	b.ResetTimer()
	for range b.N {
		if _, err := Generate(changed, first.Source); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFixtureProtectedMarkerValidation(b *testing.B) {
	result, err := Generate(acceptanceFixture(), nil)
	if err != nil {
		b.Fatal(err)
	}
	corrupted := []byte(string(result.Source) + "\n//gooo:generated:start id=\"orphan\" kind=\"activity\"\n")
	b.ResetTimer()
	for range b.N {
		if _, err := Generate(acceptanceFixture(), corrupted); err == nil {
			b.Fatal("corrupted marker was accepted")
		}
	}
}

func BenchmarkFixtureSourceMapGeneration(b *testing.B) {
	ir := acceptanceFixture()
	ir.Activities[0].Source = SourceSpan{URI: "main.gooo", Start: Position{Line: 4, Column: 1}, End: Position{Line: 9, Column: 1}}
	b.ResetTimer()
	for range b.N {
		result, err := Generate(ir, nil)
		if err != nil || len(result.SourceMap.Mappings) == 0 {
			b.Fatalf("source map generation failed: %v", err)
		}
	}
}

func BenchmarkFixtureImportPermutation(b *testing.B) {
	ir := acceptanceFixture()
	ir.Imports = []Import{{Name: "_", Path: "example/z"}, {Name: "_", Path: "example/a"}, {Name: "_", Path: "example/m"}}
	b.ResetTimer()
	for range b.N {
		if _, err := Generate(ir, nil); err != nil {
			b.Fatal(err)
		}
	}
}
