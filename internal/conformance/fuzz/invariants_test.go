package fuzz

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func assertLexResult(t *testing.T, source string, result lexResult) {
	t.Helper()
	if len(result.Tokens) == 0 {
		t.Fatal("lexer returned no tokens")
	}
	for index, token := range result.Tokens {
		assertSpan(t, source, token.Span, "token")
		if token.Lexeme != token.Text {
			t.Fatalf("token %d has mismatched text and lexeme: %#v", index, token)
		}
		if token.Text != source[token.Span.Start.Offset:token.Span.End.Offset] {
			t.Fatalf("token %d text does not match its source span: %#v", index, token)
		}
		if index < len(result.Tokens)-1 && token.Kind == syntax.TokenEOF {
			t.Fatalf("token %d is an early EOF", index)
		}
	}
	last := result.Tokens[len(result.Tokens)-1]
	if last.Kind != syntax.TokenEOF || last.Text != "" || last.Lexeme != "" {
		t.Fatalf("lexer did not end with an empty EOF token: %#v", last)
	}
	assertDiagnostics(t, source, result.Diagnostics)
}

func assertFile(t *testing.T, source string, file *syntax.File) {
	t.Helper()
	if file == nil {
		t.Fatal("parser returned a nil file")
	}
	assertSpan(t, source, file.Span, "file")
	if file.Package != nil {
		assertSpan(t, source, file.Package.Span, "package")
		assertOptionalSpan(t, source, file.Package.NameSpan, "package name")
	}
	if file.Namespace != nil {
		assertSpan(t, source, file.Namespace.Span, "namespace")
		assertOptionalSpan(t, source, file.Namespace.NameSpan, "namespace name")
	}
	if len(file.Decls) != len(file.Declarations) {
		t.Fatalf("declaration aliases differ: decls=%d declarations=%d", len(file.Decls), len(file.Declarations))
	}
	for index, declaration := range file.Declarations {
		if declaration == nil {
			t.Fatalf("declaration %d is nil", index)
		}
		assertSpan(t, source, declaration.SourceSpan(), "declaration")
		switch declaration := declaration.(type) {
		case *syntax.EntityDecl:
			assertOptionalSpan(t, source, declaration.NameSpan, "entity name")
			assertOptionalSpan(t, source, declaration.IDSpan, "entity id")
		case *syntax.ActivityDecl:
			assertOptionalSpan(t, source, declaration.NameSpan, "activity name")
			assertOptionalSpan(t, source, declaration.Result.Span, "activity result")
			if declaration.Output != declaration.Result.Name {
				t.Fatalf("activity output aliases differ: output=%q result=%q", declaration.Output, declaration.Result.Name)
			}
			if len(declaration.Inputs) != len(declaration.Parameters) {
				t.Fatalf("activity input aliases differ: inputs=%d parameters=%d", len(declaration.Inputs), len(declaration.Parameters))
			}
			for _, input := range declaration.Inputs {
				assertOptionalSpan(t, source, input.Span, "activity input")
			}
		default:
			t.Fatalf("unexpected declaration type %T", declaration)
		}
	}
}

func assertDiagnostics(t *testing.T, source string, diagnostics syntax.Diagnostics) {
	t.Helper()
	for index, diagnostic := range diagnostics {
		assertSpan(t, source, diagnostic.Span, "diagnostic")
		if diagnostic.Code == "" || diagnostic.Message == "" {
			t.Fatalf("diagnostic %d is missing a stable code or message: %#v", index, diagnostic)
		}
		if diagnostic.String() == "" {
			t.Fatalf("diagnostic %d has an empty string form", index)
		}
	}
	firstError := diagnostics.Error()
	secondError := diagnostics.Error()
	if (firstError == nil) != (secondError == nil) {
		t.Fatalf("diagnostic error presence changed between calls: %v vs %v", firstError, secondError)
	}
	if firstError != nil && firstError.Error() != secondError.Error() {
		t.Fatalf("diagnostic error text changed between calls: %q vs %q", firstError, secondError)
	}
	if diagnostics.HasErrors() != (len(diagnostics.Errors()) > 0) {
		t.Fatal("diagnostic error classification was inconsistent")
	}
}

func assertSpan(t *testing.T, source string, span syntax.Span, label string) {
	t.Helper()
	if span.Filename != fuzzFilename {
		t.Fatalf("%s has filename %q, want %q", label, span.Filename, fuzzFilename)
	}
	if span.Start.Offset < 0 || span.Start.Offset > span.End.Offset || span.End.Offset > len(source) {
		t.Fatalf("%s has out-of-bounds byte offsets: %#v for %d bytes", label, span, len(source))
	}
	if span.Start.Line < 1 || span.End.Line < span.Start.Line || span.Start.Column < 1 || span.End.Column < 1 {
		t.Fatalf("%s has invalid line/column positions: %#v", label, span)
	}
}

func assertOptionalSpan(t *testing.T, source string, span syntax.Span, label string) {
	t.Helper()
	if span.IsEmpty() {
		return
	}
	assertSpan(t, source, span, label)
}
