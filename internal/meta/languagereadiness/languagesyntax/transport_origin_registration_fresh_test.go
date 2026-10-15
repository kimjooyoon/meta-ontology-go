package languagesyntax

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func TestTransportOriginRegistrationFresh1612(t *testing.T) {
	repository := os.DirFS("../../../..")
	raw, err := os.ReadFile("../../../../examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := decodeRegistry(raw)
	if err != nil {
		t.Fatal(err)
	}
	path := "internal/meta/selfimprovementtransport/originpolicy/source.gooo"
	found := 0
	for _, item := range registry.Cases {
		if item.Path == path && item.ID == "transport-run-origin-binding" &&
			item.Scope == ScopeLanguageCapability && item.Kind == KindValid && item.OrderedCollections {
			found++
		}
	}
	if found != 1 {
		t.Fatal("source must occur once in executed corpus", found)
	}
	inventory, err := InspectSourceInventory(repository, raw)
	if err != nil || len(inventory.Unregistered) != 0 || len(inventory.Missing) != 0 {
		t.Fatal("source inventory incomplete", inventory.Unregistered, inventory.Missing, err)
	}
	result := replay.ExecuteWithOrderedCollections(repository, path, KindValid, "")
	if result.ObservedDecision != DecisionPass || !result.ASTReplayed || !result.ByteReplayed ||
		!result.SemanticReplayed || !result.GetPut || !result.PutGet {
		t.Fatalf("origin policy preservation laws: %+v", result)
	}
}
