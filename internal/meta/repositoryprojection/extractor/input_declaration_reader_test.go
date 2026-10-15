package extractor

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestInputDeclarationReaderExtractionFirst(t *testing.T) {
	result, err := ExtractWithResult("../../../..", "cmd/gooo/input_declaration_observations_test.go")
	if err != nil {
		t.Fatal("source observation reader cannot be projected", err)
	}
	if len(result.Generated) < 2 {
		t.Fatal("expected complete-declaration partitions")
	}
	for name, raw := range result.Generated {
		if physicalLines(raw) > 75 {
			t.Fatal("partition exceeds current source profile", name)
		}
		if _, err := parser.ParseFile(token.NewFileSet(), name, raw, parser.AllErrors); err != nil {
			t.Fatal(name, err)
		}
	}
}
