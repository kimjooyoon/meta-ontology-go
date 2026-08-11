package syntax

import (
	"reflect"
	"testing"
)

func TestLexTokenKindsTable(t *testing.T) {
	tests := []struct {
		name   string
		source string
		kinds  []TokenKind
		texts  []string
	}{
		{
			name: "billing declarations",
			source: `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Order, PaymentMethod) -> Payment`,
			kinds: []TokenKind{
				TokenPackage, TokenIdentifier, TokenNamespace, TokenIdentifier,
				TokenEntity, TokenIdentifier, TokenID, TokenString,
				TokenActivity, TokenIdentifier, TokenLParen, TokenIdentifier,
				TokenComma, TokenIdentifier, TokenRParen, TokenArrow, TokenIdentifier,
				TokenEOF,
			},
		},
		{
			name:   "comments and string escapes",
			source: "/* leading */ entity\tThing id \"urn:test\\nvalue\" // trailing\n",
			kinds:  []TokenKind{TokenEntity, TokenIdentifier, TokenID, TokenString, TokenEOF},
			texts:  []string{"entity", "Thing", "id", `"urn:test\nvalue"`, ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens, diagnostics := Lex(test.source)
			if len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %v", diagnostics)
			}
			gotKinds := make([]TokenKind, len(tokens))
			for i, token := range tokens {
				gotKinds[i] = token.Kind
			}
			if !reflect.DeepEqual(gotKinds, test.kinds) {
				t.Fatalf("token kinds = %v, want %v", gotKinds, test.kinds)
			}
			if test.texts != nil {
				gotTexts := make([]string, len(tokens))
				for i, token := range tokens {
					gotTexts[i] = token.Text
				}
				if !reflect.DeepEqual(gotTexts, test.texts) {
					t.Fatalf("token texts = %q, want %q", gotTexts, test.texts)
				}
				if tokens[3].Value != "urn:test\nvalue" {
					t.Fatalf("decoded string = %q", tokens[3].Value)
				}
			}
		})
	}
}

func TestLexSpansAndDeterminism(t *testing.T) {
	source := "package p\nnamespace n\n"
	firstTokens, firstDiagnostics := LexFile("billing.gooo", source)
	secondTokens, secondDiagnostics := LexFile("billing.gooo", source)
	if !reflect.DeepEqual(firstTokens, secondTokens) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Fatal("lexing the same source was not deterministic")
	}

	if got, want := firstTokens[0].Span, (Span{Filename: "billing.gooo", Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 7, Line: 1, Column: 8}}); got != want {
		t.Fatalf("package span = %#v, want %#v", got, want)
	}
	if got, want := firstTokens[2].Span.Start, (Position{Offset: 10, Line: 2, Column: 1}); got != want {
		t.Fatalf("namespace start = %#v, want %#v", got, want)
	}
	if got, want := firstTokens[len(firstTokens)-1].Span.Start, (Position{Offset: len(source), Line: 3, Column: 1}); got != want {
		t.Fatalf("EOF position = %#v, want %#v", got, want)
	}
}

func TestLexDiagnosticsTable(t *testing.T) {
	tests := []struct {
		name   string
		source string
		code   DiagnosticCode
	}{
		{name: "bad character", source: "entity A id \"x\" @", code: DiagUnexpectedCharacter},
		{name: "unterminated block comment", source: "/* comment", code: DiagUnterminatedComment},
		{name: "unterminated string", source: "entity A id \"x", code: DiagUnterminatedString},
		{name: "invalid escape", source: "entity A id \"\\q\"", code: DiagInvalidEscape},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := Lex(test.source)
			if len(diagnostics) != 1 || diagnostics[0].Code != test.code {
				t.Fatalf("diagnostics = %#v, want one %q", diagnostics, test.code)
			}
		})
	}
}

func TestParseValidTable(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		packageName string
		namespace   string
		entities    int
		activities  int
	}{
		{
			name: "billing example",
			source: `package billing
namespace billing
entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"
activity PayOrder(Order, PaymentMethod) -> Payment`,
			packageName: "billing", namespace: "billing", entities: 3, activities: 1,
		},
		{
			name: "empty parameter list",
			source: `package p
namespace n
activity Tick() -> Result`,
			packageName: "p", namespace: "n", activities: 1,
		},
		{
			name: "unicode identifiers",
			source: `package 도메인
namespace 도메인
entity 주문 id "urn:order"
activity 결제(주문) -> 주문`,
			packageName: "도메인", namespace: "도메인", entities: 1, activities: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, diagnostics := Parse(test.source)
			if len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %v", diagnostics)
			}
			if file.Package == nil || file.Package.Name != test.packageName || file.Namespace == nil || file.Namespace.Name != test.namespace {
				t.Fatalf("headers = %#v", file)
			}
			if len(file.Declarations) != test.entities+test.activities || len(file.Decls) != len(file.Declarations) {
				t.Fatalf("declarations = %#v", file.Declarations)
			}
			for _, declaration := range file.Declarations {
				switch declaration.(type) {
				case *EntityDecl:
					if test.entities == 0 {
						t.Fatalf("unexpected entity declaration")
					}
					test.entities--
				case *ActivityDecl:
					if test.activities == 0 {
						t.Fatalf("unexpected activity declaration")
					}
					test.activities--
				default:
					t.Fatalf("unexpected declaration type %T", declaration)
				}
			}
			if test.entities != 0 || test.activities != 0 {
				t.Fatalf("declaration counts were not preserved")
			}
		})
	}
}

func TestParseDiagnosticsTable(t *testing.T) {
	tests := []struct {
		name   string
		source string
		codes  []DiagnosticCode
	}{
		{
			name:   "missing headers",
			source: "",
			codes:  []DiagnosticCode{DiagExpectedPackage, DiagExpectedNamespace},
		},
		{
			name:   "missing entity id and string",
			source: "package p namespace n entity Thing",
			codes:  []DiagnosticCode{DiagExpectedID, DiagExpectedString},
		},
		{
			name:   "missing parameter comma",
			source: "package p namespace n activity A(One Two) -> Result",
			codes:  []DiagnosticCode{DiagExpectedComma},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := Parse(test.source)
			got := make([]DiagnosticCode, len(diagnostics))
			for i, diagnostic := range diagnostics {
				got[i] = diagnostic.Code
			}
			if !reflect.DeepEqual(got, test.codes) {
				t.Fatalf("diagnostic codes = %v, want %v (%v)", got, test.codes, diagnostics)
			}
		})
	}
}

func TestParsePreservesDeclarationSpans(t *testing.T) {
	source := "package p\nnamespace n\nentity Order id \"urn:order\"\nactivity Pay(Order) -> Order\n"
	file, diagnostics := ParseFile("billing.gooo", source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	entity := file.Declarations[0].(*EntityDecl)
	activity := file.Declarations[1].(*ActivityDecl)
	if entity.Span.Filename != "billing.gooo" || entity.Span.Start.Line != 3 || entity.Span.End.Line != 3 {
		t.Fatalf("entity span = %#v", entity.Span)
	}
	if activity.Span.Start.Line != 4 || activity.Span.End.Line != 4 || activity.Parameters[0].Span.Start.Line != 4 {
		t.Fatalf("activity spans = %#v", activity)
	}
}
