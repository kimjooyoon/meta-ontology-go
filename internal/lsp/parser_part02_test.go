package lsp

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"testing"
)

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
