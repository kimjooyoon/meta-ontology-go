package languagesyntax

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func TestEverydayRegistryRetainsAllPriorCases67143(t *testing.T) {
	raw := everydayRegistryRead67143(t, "examples/language-syntax-roundtrip/corpus.json")
	registry, err := decodeRegistry(raw)
	if err != nil || FixedTotal < 111 || FixedCapabilityTotal+FixedGovernanceTotal != FixedTotal || FixedGovernanceTotal != 2 {
		t.Fatal("current exact registry", err)
	}
	archive := everydayRegistryRead67143(t, "docs/observations/everyday-registration/prior-corpus.json.gz")
	reader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	previous, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var prior Registry
	if err := json.Unmarshal(previous, &prior); err != nil {
		t.Fatal(err)
	}
	if len(prior.Cases) != 110 || len(registry.Cases) != FixedTotal ||
		!reflect.DeepEqual(prior.Cases, registry.Cases[:110]) ||
		!reflect.DeepEqual(prior.MetaSources, registry.MetaSources) ||
		!reflect.DeepEqual(prior.ProjectionUnits, registry.ProjectionUnits) ||
		!reflect.DeepEqual(prior.PackageUnits, registry.PackageUnits) {
		t.Fatal("previous source identities, scopes or classifications changed")
	}
	added := registry.Cases[110]
	if added.ID != "everyday-local-condition-origins" || added.Path != "examples/everyday-program/main.gooo" ||
		added.Kind != KindValid || added.Scope != ScopeLanguageCapability || added.ExpectedDecision != DecisionPass {
		t.Fatal("ordinary source registration omitted", added)
	}
}

func TestEverydayRegistryCoversActualSourceInventory67143(t *testing.T) {
	registry, err := decodeRegistry(everydayRegistryRead67143(t, "examples/language-syntax-roundtrip/corpus.json"))
	if err != nil {
		t.Fatal(err)
	}
	observed, err := replay.Observe(os.DirFS("../../../.."))
	if err != nil {
		t.Fatal(err)
	}
	extra, missing := compareCorpus(registry, observed)
	if len(extra) != 0 || len(missing) != 0 {
		t.Fatal("complete source inventory", extra, missing)
	}
	for i, definition := range registry.Cases {
		if definition.Path != "examples/everyday-program/main.gooo" {
			continue
		}
		registry.Cases = append(registry.Cases[:i:i], registry.Cases[i+1:]...)
		unregistered, _ := compareCorpus(registry, observed)
		if len(unregistered) != 1 || unregistered[0] != definition.Path {
			t.Fatal("omission counterexample was not detected", unregistered)
		}
		return
	}
	t.Fatal("registered example missing")
}

func everydayRegistryRead67143(t *testing.T, path string) []byte {
	t.Helper()
	root := os.Getenv("GOOO_EVERYDAY_REGISTRY_ROOT67143")
	if root == "" {
		root = "../../../.."
	}
	raw, err := os.ReadFile(root + "/" + path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
