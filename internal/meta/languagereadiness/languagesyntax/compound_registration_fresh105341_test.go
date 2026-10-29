package languagesyntax

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func TestCompoundSourceRegistration105341(t *testing.T) {
	raw, err := os.ReadFile("../../../../examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := decodeRegistry(raw)
	if err != nil || len(registry.Cases) != FixedTotal || FixedCapabilityTotal+FixedGovernanceTotal != FixedTotal || FixedGovernanceTotal != 2 {
		t.Fatalf("current registry: cases=%d language=%d governance=%d: %v", len(registry.Cases), FixedCapabilityTotal, FixedGovernanceTotal, err)
	}
	previous, err := os.ReadFile("../../../../docs/observations/compound-updates/prior-corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	var prior Registry
	if err = json.Unmarshal(previous, &prior); err != nil {
		t.Fatal(err)
	}
	// The v74 population is historical; later registrations may append cases.
	if len(registry.Cases) <= len(prior.Cases) {
		t.Fatalf("compound registration missing after %d prior cases", len(prior.Cases))
	}
	if !historicalSyntaxCasesRetained(prior.Cases, registry.Cases) {
		t.Fatal("prior source identities changed")
	}
	definition := registry.Cases[len(prior.Cases)]
	if definition.ID != "compound-local-update-origins" || definition.Path != "examples/compound-updates/main.gooo" || definition.Scope != ScopeLanguageCapability {
		t.Fatal("source owner missing", definition)
	}
	observed, err := replay.Observe(os.DirFS("../../../.."))
	if err != nil {
		t.Fatal(err)
	}
	extra, missing := compareCorpus(registry, observed)
	if len(extra) != 0 || len(missing) != 0 {
		t.Fatal("current source inventory", extra, missing)
	}
}
