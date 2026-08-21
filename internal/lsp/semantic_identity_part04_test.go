package lsp

import (
	"context"
	"testing"
)

func TestLoweredIdentityRenameAndNamespaceLocality(t *testing.T) {
	first := parseIdentityResult(t, `package billing
namespace billing
entity Order id "billing://entity/order"
`)
	second := parseIdentityResult(t, `package billing
namespace billing
entity Purchase id "billing://entity/order"
`)
	if first.Symbols[0].ID != second.Symbols[0].ID {
		t.Fatalf("explicit-ID rename changed identity: %q -> %q", first.Symbols[0].ID, second.Symbols[0].ID)
	}
	billing := parseIdentityResult(t, `package billing
namespace billing
entity Order id "billing://entity/order"
`)
	settlement := parseIdentityResult(t, `package settlement
namespace settlement
entity Order id "settlement://entity/order"
`)
	if billing.Symbols[0].ID == settlement.Symbols[0].ID {
		t.Fatalf("namespace-local declarations collapsed: %#v / %#v", billing.Symbols, settlement.Symbols)
	}
}
func TestMalformedOrUnknownLoweringHasDiagnosticsAndNoLinks(t *testing.T) {
	result, err := (SyntaxParser{}).ParseContext(context.Background(), "unknown.gooo", `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Missing) -> Order
`)
	if err != nil || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "semantic.lowering" {
		t.Fatalf("unknown declaration diagnostics = %#v, error = %v", result.Diagnostics, err)
	}
	for _, symbol := range result.Symbols {
		if symbol.ID != "" || symbol.hasIdentity {
			t.Fatalf("unknown declaration invented symbol identity: %#v", symbol)
		}
	}
	for _, reference := range result.References {
		if reference.ID != "" {
			t.Fatalf("unknown declaration invented reference identity: %#v", reference)
		}
	}
}
func parseIdentityResult(t *testing.T, source string) ParseResult {
	t.Helper()
	result, err := (SyntaxParser{}).ParseContext(context.Background(), "identity.gooo", source)
	if err != nil || len(result.Diagnostics) != 0 || !result.semanticValid {
		t.Fatalf("identity parse = %#v, error = %v", result, err)
	}
	return result
}
func hasDocumentSymbol(symbols []DocumentSymbol, name string, kind SymbolKind) bool {
	for _, symbol := range symbols {
		if symbol.Name == name && symbol.Kind == kind {
			return true
		}
	}
	return false
}
