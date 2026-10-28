package languagesyntax

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"slices"
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
		!metaSourceHistoryRetainedByCases67143(prior, registry) ||
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

// Existing source identities retain their ordered prefix; new meta sources are
// admitted separately by decodeRegistry's exact current native/JSON binding.
func metaSourceHistoryRetained67143(prior, current []string) bool {
	return len(current) >= len(prior) && reflect.DeepEqual(prior, current[:len(prior)])
}

// A raw meta source can become a newly appended executable capability case.
// Its path remains registered, and every other raw source keeps its order.
func metaSourceHistoryRetainedByCases67143(prior, current Registry) bool {
	if len(current.Cases) < len(prior.Cases) || !reflect.DeepEqual(prior.Cases, current.Cases[:len(prior.Cases)]) {
		return false
	}
	promoted := map[string]bool{}
	for _, definition := range current.Cases[len(prior.Cases):] {
		if definition.Kind == KindValid && definition.ExpectedDecision == DecisionPass &&
			definition.Scope == ScopeLanguageCapability && definition.MetaOperation == "replay-language-syntax" {
			promoted[definition.Path] = true
		}
	}
	retained := []string{}
	for _, path := range prior.MetaSources {
		if promoted[path] {
			if slices.Contains(current.MetaSources, path) {
				return false
			}
		} else {
			retained = append(retained, path)
		}
	}
	return metaSourceHistoryRetained67143(retained, current.MetaSources)
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
