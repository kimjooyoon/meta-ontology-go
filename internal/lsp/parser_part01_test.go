package lsp

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"testing"
)

func TestSyntaxParserUsesCanonicalASTAndSpans(t *testing.T) {
	source := "package billing\r\nnamespace billing\r\nentity Order id \"billing://entity/order\"\r\nactivity Pay(Order) -> Order"
	first, err := (SyntaxParser{}).ParseContext(context.Background(), "billing.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	second, err := (SyntaxParser{}).ParseContext(context.Background(), "billing.gooo", source)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("parse is not deterministic: first=%#v second=%#v error=%v", first, second, err)
	}
	if first.File == nil || len(first.File.Decls) != 2 || len(first.Diagnostics) != 0 {
		t.Fatalf("parse result = %#v", first)
	}
	entity := first.Symbols[0]
	if entity.Name != "Order" || entity.ID != "billing://entity/order" || entity.SelectionRange.Start != (Position{Line: 2, Character: 7}) {
		t.Fatalf("entity symbol = %#v", entity)
	}
	if len(first.References) != 2 || first.References[0].Range.Start != (Position{Line: 3, Character: 13}) {
		t.Fatalf("references = %#v", first.References)
	}
}
func TestSyntaxDiagnosticsSortByCanonicalSourceOrder(t *testing.T) {
	source := "package namespace n\n@"
	_, raw := syntax.ParseFile("fixture.gooo", source)
	if len(raw) != 2 || raw[0].Code != syntax.DiagUnexpectedCharacter || raw[1].Code != syntax.DiagExpectedIdentifier {
		t.Fatalf("raw phase order = %#v", raw)
	}
	canonical := raw.SortBySpan()
	if canonical[0].Code != syntax.DiagExpectedIdentifier || canonical[1].Code != syntax.DiagUnexpectedCharacter {
		t.Fatalf("canonical source order = %#v", canonical)
	}
	first, err := (SyntaxParser{}).ParseContext(context.Background(), "fixture.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	second, err := (SyntaxParser{}).ParseContext(context.Background(), "fixture.gooo", source)
	if err != nil || !reflect.DeepEqual(first.Diagnostics, second.Diagnostics) {
		t.Fatalf("repeated diagnostics differ: first=%#v second=%#v error=%v", first.Diagnostics, second.Diagnostics, err)
	}
	want := []string{"parse.expected-identifier", "lex.unexpected-character"}
	for index, code := range want {
		if first.Diagnostics[index].Code != code {
			t.Fatalf("sorted diagnostics = %#v, want codes %v", first.Diagnostics, want)
		}
	}
	permuted := append([]Diagnostic(nil), first.Diagnostics...)
	permuted[0], permuted[1] = permuted[1], permuted[0]
	if sorted := canonicalDiagnosticOrder("fixture.gooo", source, permuted); !reflect.DeepEqual(first.Diagnostics, sorted) {
		t.Fatalf("permuted diagnostics = %#v, want %#v", sorted, first.Diagnostics)
	}
}
