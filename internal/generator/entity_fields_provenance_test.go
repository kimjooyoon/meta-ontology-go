package generator

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestEntityFieldsProvenanceFailuresAreTransactional(t *testing.T) {
	cases := []struct {
		name string
		edit func(*SemanticIR)
		code string
	}{
		{name: "optional", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].Presence = "optional" }, code: entityFieldsUnsupportedShapeDiagnostic},
		{name: "many", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].Cardinality = "many" }, code: entityFieldsUnsupportedShapeDiagnostic},
		{name: "unsupported type", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].TypeRefID = "urn:gooo:type:integer" }, code: entityFieldsUnsupportedTypeDiagnostic},
		{name: "missing parent", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].Parent = "" }, code: entityFieldsIncompleteDiagnostic},
		{name: "wrong parent", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].Parent = "urn:gooo:entity:other" }, code: entityFieldsWrongParentDiagnostic},
		{name: "duplicate field ID", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[1].ID = ir.Entities[0].Fields[0].ID }, code: entityFieldsIDCollisionDiagnostic},
		{name: "cross-kind field ID", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].ID = ir.Activities[0].ID }, code: entityFieldsIDCollisionDiagnostic},
		{name: "illegal declaration order", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[1].Source.Start.Offset = 10 }, code: entityFieldsIllegalReorderDiagnostic},
		{name: "zero origin", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].Source = SourceSpan{} }, code: entityFieldsIncompleteDiagnostic},
		{name: "missing ID span", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].IDSpan = SourceSpan{} }, code: entityFieldsIncompleteDiagnostic},
		{name: "missing name span", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].NameSpan = SourceSpan{} }, code: entityFieldsIncompleteDiagnostic},
		{name: "missing type span", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].TypeRefSpan = SourceSpan{} }, code: entityFieldsIncompleteDiagnostic},
		{name: "missing presence span", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].PresenceSpan = SourceSpan{} }, code: entityFieldsIncompleteDiagnostic},
		{name: "missing cardinality span", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].CardinalitySpan = SourceSpan{} }, code: entityFieldsIncompleteDiagnostic},
		{name: "cross snapshot", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[1].Source.URI = "other.gooo" }, code: entityFieldsUnrepresentableDiagnostic},
		{name: "wrong subspan snapshot", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].IDSpan.URI = "other.gooo" }, code: entityFieldsUnrepresentableDiagnostic},
		{name: "subspan outside parent", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].NameSpan.Start.Offset = 19 }, code: entityFieldsUnrepresentableDiagnostic},
		{name: "overlapping subspans", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].NameSpan.End.Offset = 27 }, code: entityFieldsIllegalReorderDiagnostic},
		{name: "illegal subspan order", edit: func(ir *SemanticIR) {
			ir.Entities[0].Fields[0].TypeRefSpan.Start.Offset = 23
			ir.Entities[0].Fields[0].TypeRefSpan.End.Offset = 25
		}, code: entityFieldsIllegalReorderDiagnostic},
		{name: "ambiguous name provenance", edit: func(ir *SemanticIR) {
			ir.Entities[0].Fields[0].NameSource = SourceSpan{URI: "other.gooo", Start: Position{Offset: 25, Line: 4, Column: 10}, End: Position{Offset: 27, Line: 4, Column: 12}}
		}, code: entityFieldsUnrepresentableDiagnostic},
		{name: "Go name collision", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[1].Name = ir.Entities[0].Fields[0].Name }, code: entityFieldsGoNameCollisionDiagnostic},
		{name: "aliases", edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].Aliases = []string{"OrderNo"} }, code: entityFieldsUnsupportedShapeDiagnostic},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ir := entityFieldsFixture()
			testCase.edit(&ir)
			root := t.TempDir()
			sentinel := root + "/sentinel"
			if err := os.WriteFile(sentinel, []byte("unchanged"), 0o600); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(sentinel)
			if err != nil {
				t.Fatal(err)
			}
			result, err := New(Options{}).generateWithEntityFieldsSupport(ir, nil, supportedEntityFieldsForTest())
			if err == nil || !strings.Contains(err.Error(), testCase.code) || (!strings.Contains(err.Error(), ".gooo:") && testCase.name != "zero origin") {
				t.Fatalf("expected deterministic source-backed %s, got result=%#v err=%v", testCase.code, result, err)
			}
			if result.Source != nil || result.SourceMap.Mappings != nil {
				t.Fatalf("rejected input produced artifacts: %#v", result)
			}
			projection, projectionErr := generateProjectionV1WithEntityFieldsSupport(New(Options{}), ir, nil, supportedEntityFieldsForTest())
			if projectionErr == nil || projection.Schema != "" || projection.Source != nil || projection.SourceMap.Mappings != nil || projection.SemanticIR.Entities != nil || projection.Metadata.SourceDigest != "" {
				t.Fatalf("rejected projection produced evidence: %#v err=%v", projection, projectionErr)
			}
			after, err := os.ReadFile(sentinel)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("rejected generation changed filesystem: before=%q after=%q err=%v", before, after, err)
			}
		})
	}
}
