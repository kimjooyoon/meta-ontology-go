package analyzer

import (
	"strings"
	"testing"
)

func TestDuplicateRegistryRegistrationIsNoMutation(t *testing.T) {
	registry := NewRegistry()
	entry := Registration{
		Ref:  SymbolRef{PackagePath: "billing", PackageName: "billing", Name: "Order"},
		Kind: KindEntity, Identity: NewIdentity("billing", "billing://entity/order"),
	}
	if err := registry.Register(entry); err != nil {
		t.Fatal(err)
	}
	canonical := registry.Canonical()
	digest := registry.Digest()
	duplicate := entry
	duplicate.Span = Span{Filename: "different.go", Start: Position{Offset: 9}, End: Position{Offset: 12}}
	if err := registry.Register(duplicate); err != nil {
		t.Fatal(err)
	}
	if registry.Canonical() != canonical || registry.Digest() != digest || len(registry.all()) != 1 {
		t.Fatal("duplicate registration changed registry identity")
	}
}
func TestRegistryRejectsMalformedIdentityAndNamespaceWithoutMutation(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		identity    Identity
		wantMessage string
	}{
		{name: "invalid identity", identity: NewIdentity("billing", "Order"), wantMessage: "semantic identity"},
		{name: "empty namespace", identity: NewIdentity("", "billing://entity/order"), wantMessage: "semantic namespace"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			registry := NewRegistry()
			beforeCanonical, beforeDigest := registry.Canonical(), registry.Digest()
			err := registry.Register(Registration{
				Ref:  SymbolRef{PackagePath: "billing", PackageName: "billing", Name: "Order"},
				Kind: KindEntity, Identity: testCase.identity,
			})
			if err == nil || !strings.Contains(err.Error(), testCase.wantMessage) {
				t.Fatalf("malformed registration error = %v, want %q", err, testCase.wantMessage)
			}
			if registry.Canonical() != beforeCanonical || registry.Digest() != beforeDigest || len(registry.all()) != 0 {
				t.Fatal("malformed registration changed registry state")
			}
		})
	}
}
