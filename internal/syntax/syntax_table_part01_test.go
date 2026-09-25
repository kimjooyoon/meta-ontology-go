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
