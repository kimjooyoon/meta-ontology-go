package lsp

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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

func TestCanonicalASTAliasesCannotDiverge(t *testing.T) {
	source := "package p\nnamespace n\nentity A id \"urn:a\"\nactivity Run(A) -> A"
	file, diagnostics := syntax.ParseFile("alias.gooo", source)
	canonical, err := adaptSyntaxResult("alias.gooo", source, file, diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	activity := *file.Decls[1].(*syntax.ActivityDecl)
	activity.Parameters = []syntax.NameRef{{Name: "alias-input", Span: activity.Parameters[0].Span}}
	activity.Result = syntax.NameRef{Name: "alias-output", Span: activity.Result.Span}
	variantFile := *file
	variantFile.Decls = []syntax.Declaration{file.Decls[0], &activity}
	variantFile.Declarations = []syntax.Declaration{&syntax.EntityDecl{Name: "alias-only"}}
	variant, err := adaptSyntaxResult("alias.gooo", source, &variantFile, diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(canonical.Symbols, variant.Symbols) || !reflect.DeepEqual(canonical.References, variant.References) {
		t.Fatalf("non-preferred alias fields changed LSP output: canonical=%#v/%#v variant=%#v/%#v", canonical.Symbols, canonical.References, variant.Symbols, variant.References)
	}
}

func TestDocumentStoresImmutableLSPProjectionSnapshot(t *testing.T) {
	uri := "file:///immutable.gooo"
	symbols := []Symbol{{Name: "Order", ID: "urn:order", Kind: SymbolClass}}
	references := []Reference{{Name: "Order"}}
	diagnostics := []Diagnostic{{Message: "original"}}
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{Symbols: symbols, References: references, Diagnostics: diagnostics}
	})
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "Order"},
	})
	writeRequest(t, &input, 1, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	server := NewServer(parser)
	if err := server.Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}

	symbols[0].Name = "mutated"
	references[0].Name = "mutated"
	diagnostics[0].Message = "mutated"
	document := server.documents[uri]
	if document == nil {
		t.Fatal("document was not stored")
	}
	if document.result.Symbols[0].Name != "Order" || document.result.References[0].Name != "Order" ||
		document.result.Diagnostics[0].Message != "original" {
		t.Fatalf("stored LSP projection changed after parser reuse: %#v", document.result)
	}
	snapshot := documentCopy(document)
	snapshot.result.Symbols[0].Name = "reader mutation"
	snapshot.result.References[0].Name = "reader mutation"
	snapshot.result.Diagnostics[0].Message = "reader mutation"
	if document.result.Symbols[0].Name != "Order" || document.result.References[0].Name != "Order" ||
		document.result.Diagnostics[0].Message != "original" {
		t.Fatalf("feature snapshot mutation changed stored projection: %#v", document.result)
	}
}

func TestSyntaxDiagnosticsPreserveEOFAndCRLFSpans(t *testing.T) {
	source := "package billing\r\nnamespace billing\r\nentity Order id \"unterminated"
	result, err := (SyntaxParser{}).ParseContext(context.Background(), "billing.gooo", source)
	if err != nil || len(result.Diagnostics) == 0 {
		t.Fatalf("diagnostics = %#v, error = %v", result.Diagnostics, err)
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.Code != "lex.unterminated-string" || diagnostic.Source != "gooo" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	if diagnostic.Range.Start != (Position{Line: 2, Character: 16}) || diagnostic.Range.End.Character <= 16 {
		t.Fatalf("diagnostic range = %#v", diagnostic.Range)
	}
}

func TestSyntaxDiagnosticsHandleInvalidUTF8Boundary(t *testing.T) {
	bytes := []byte("package p\nnamespace n\nentity A id \"x\" ")
	bytes = append(bytes, 0xff)
	result, err := (SyntaxParser{}).ParseContext(context.Background(), "invalid.gooo", string(bytes))
	if err != nil || len(result.Diagnostics) == 0 {
		t.Fatalf("diagnostics = %#v, error = %v", result.Diagnostics, err)
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.Code != "lex.unexpected-character" || diagnostic.Range.Start.Character != 16 {
		t.Fatalf("invalid byte diagnostic = %#v", diagnostic)
	}
}

func TestSyntaxEntityIdentitySpanHandlesInvalidUTF8WithoutMutation(t *testing.T) {
	bytes := []byte("package p\r\nnamespace n\r\nentity Order id \"urn:order\"\r\n")
	bytes = append(bytes, 0xff)
	source := string(bytes)
	file, diagnostics := syntax.ParseFile("invalid-identity.gooo", source)
	originalSource := source
	originalDiagnostics := append(syntax.Diagnostics(nil), diagnostics...)
	result, err := adaptSyntaxResult("invalid-identity.gooo", source, file, diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Symbols) != 1 {
		t.Fatalf("symbols = %#v", result.Symbols)
	}
	symbol := result.Symbols[0]
	if symbol.hasIdentity || symbol.ID != "" {
		t.Fatalf("malformed source invented identity = %#v", symbol)
	}
	if source != originalSource || !reflect.DeepEqual(diagnostics, originalDiagnostics) {
		t.Fatalf("identity mapping mutated source or diagnostics: source=%q diagnostics=%#v", source, diagnostics)
	}
}
