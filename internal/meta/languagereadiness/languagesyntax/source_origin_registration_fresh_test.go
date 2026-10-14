package languagesyntax

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func TestOriginSourceRegistration27537(t *testing.T) {
	snapshot := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	repository := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if repository == "" {
		repository = "../../../.."
	}
	read := func(path string) []byte {
		t.Helper()
		if snapshot != "" {
			if raw, err := os.ReadFile(filepath.Join(snapshot, path)); err == nil {
				return raw
			}
		}
		raw, err := os.ReadFile(filepath.Join(repository, path))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	registry, err := decodeRegistry(read("examples/language-syntax-roundtrip/corpus.json"))
	if err != nil || len(registry.Cases) != 110 || FixedCapabilityTotal != 108 || FixedGovernanceTotal != 2 {
		t.Fatal("exact complete registry", registry, err)
	}
	paths := []string{"examples/fill-candidate-origins/source.gooo", "examples/fill-candidate-origins/inspect.gooo",
		"examples/derived-candidate-origins/source.gooo", "examples/derived-candidate-origins/inspect.gooo"}
	files := fstest.MapFS{}
	for _, path := range paths {
		found := false
		for _, item := range registry.Cases {
			if item.Path == path {
				found = item.Kind == KindValid && item.Scope == ScopeLanguageCapability && item.OrderedCollections
			}
		}
		if !found {
			t.Fatal("real Gooo source omitted or reclassified", path)
		}
		files[path] = &fstest.MapFile{Data: read(path)}
	}
	for _, path := range paths {
		r := replay.ExecuteWithOrderedCollections(files, path, KindValid, "")
		if r.ObservedDecision != DecisionPass || !r.ASTReplayed || !r.ByteReplayed || !r.SemanticReplayed || !r.GetPut || !r.PutGet {
			t.Fatal("new registered source failed preservation laws", path, r)
		}
	}
	t.Log("registration27537:all4 real source/reader programs registered as language capability;4 syntax/AST/byte/semantic/BX replays pass;corpus110 valid107 invalid3 language108 governance2;no producer/model calls")
}
