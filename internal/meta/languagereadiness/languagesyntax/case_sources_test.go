package languagesyntax

import (
	"io/fs"
	"os"
	"reflect"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestRegisteredSourceFamilyRetainsCapabilityAndFileDenominators(t *testing.T) {
	repository := os.DirFS("../../../..")
	raw, err := fs.ReadFile(repository, "examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := decodeRegistry(raw)
	if err != nil || len(registry.CaseSources) != 1 || len(registry.Cases) != FixedTotal {
		t.Fatal("physical grouping changed capability population", err)
	}
	if FixedCapabilityTotal+FixedGovernanceTotal != FixedTotal {
		t.Fatal("capability and governance populations do not partition the registry")
	}
	prior := registry
	prior.CaseSources = nil
	if !reflect.DeepEqual(registryPaths(registry), registryPaths(prior)) {
		t.Fatal("already registered family members changed the physical file population")
	}
	var definition CaseDefinition
	for _, item := range registry.Cases {
		if item.ID == "explicit-consumer-continuation" {
			definition = item
		}
	}
	names := caseSourcesFor(registry, definition.ID)
	result := ReplayCaseSources(repository, definition, names)
	if result.ObservedDecision != DecisionPass || len(result.SourceMembers) != 2 || result.SourceLines != 91 {
		t.Fatal("actual continuation family was not jointly replayed", result)
	}
	legacy := ReplayCase(repository, definition)
	if legacy.ObservedDecision != DecisionPass || result.SemanticDigest == legacy.SemanticDigest {
		t.Fatal("additional recovery declarations were not included in the semantic image")
	}
	cloned := cloneCaseSources(registry.CaseSources)
	cloned[0].Members[0] = "changed.gooo"
	if registry.CaseSources[0].Members[0] == "changed.gooo" {
		t.Fatal("report membership aliases its registry")
	}
}

func TestCaseSourceSetsRetainTheDeclaredProfile(t *testing.T) {
	files := fstest.MapFS{
		"types.gooo": &fstest.MapFile{Data: []byte("package p\nnamespace p\nentity Items id \"urn:items\" fields { field values id \"urn:values\" type integer required many }\n")},
		"run.gooo":   &fstest.MapFile{Data: []byte("package p\nnamespace p\nactivity Read(Items) -> Items id \"urn:read\" computes \"return input\"\n")},
	}
	definition := CaseDefinition{ID: "records", Path: "types.gooo", Kind: KindValid,
		EntityFieldsProfile: syntax.EntityFieldsV5Support().Profile.ID}
	names := []string{"types.gooo", "run.gooo"}
	result := ReplayCaseSources(files, definition, names)
	if result.ObservedDecision != DecisionPass || result.EntityFieldsProfile != definition.EntityFieldsProfile || result.ProfileDigest == "" {
		t.Fatal("explicit group profile was not retained", result)
	}
	definition.EntityFieldsProfile = syntax.EntityFieldsV1Support().Profile.ID
	if result := ReplayCaseSources(files, definition, names); result.ObservedDecision == DecisionPass {
		t.Fatal("grouping silently widened V1", result)
	}
	definition.EntityFieldsProfile = "unknown-profile"
	if result := ReplayCaseSources(files, definition, names); result.ObservedDecision != replay.DecisionUnknown {
		t.Fatal("unknown group profile was guessed", result)
	}
}

func TestCaseSourceSetMembershipCannotReplaceACaseOrItsPrimary(t *testing.T) {
	for _, mode := range []string{"unknown-case", "duplicate-case", "primary", "duplicate-member", "invalid-case", "single", "alias"} {
		registry := expectedRegistry()
		group := &registry.CaseSources[0]
		switch mode {
		case "unknown-case":
			group.CaseID = "unknown"
		case "duplicate-case":
			registry.CaseSources = append(registry.CaseSources, *group)
		case "primary":
			slices.Reverse(group.Members)
		case "duplicate-member":
			group.Members[1] = group.Members[0]
		case "invalid-case":
			for i := range registry.Cases {
				if registry.Cases[i].ID == group.CaseID {
					registry.Cases[i].Kind = KindInvalid
				}
			}
		case "single":
			group.Members = group.Members[:1]
		case "alias":
			group.Members[1] = "./other.gooo"
		}
		if validateCaseSources(registry) == nil {
			t.Fatal("invalid source membership accepted", mode)
		}
	}
}

func TestCaseSourceMembershipRequiresExactIdentityAndReturnsOwnedSlices(t *testing.T) {
	registry := expectedRegistry()
	var definition CaseDefinition
	for _, item := range registry.Cases {
		if item.ID == registry.CaseSources[0].CaseID {
			definition = item
		}
	}
	names, err := CaseSourceMembers(registry, definition)
	if err != nil || len(names) != 2 {
		t.Fatal(err)
	}
	names[1] = "changed.gooo"
	if registry.CaseSources[0].Members[1] == names[1] {
		t.Fatal("caller owns the registry's slice")
	}
	changed := definition
	changed.EntityFieldsProfile = "unselected"
	if _, err := CaseSourceMembers(registry, changed); err == nil {
		t.Fatal("another profile inherited this case's source membership")
	}
	registry.Cases = append(registry.Cases, definition)
	if _, err := CaseSourceMembers(registry, definition); err == nil {
		t.Fatal("duplicate case identity accepted")
	}
}
