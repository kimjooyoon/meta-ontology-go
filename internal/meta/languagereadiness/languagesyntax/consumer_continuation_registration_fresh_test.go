package languagesyntax

import (
	"os"
	"slices"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestConsumerContinuationRegistrationFresh1619(t *testing.T) {
	repository := os.DirFS("../../../..")
	raw, err := os.ReadFile("../../../../examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := decodeRegistry(raw)
	if err != nil {
		t.Fatal(err)
	}
	path := "internal/meta/executioncontinuation/policy/source.gooo"
	found := 0
	for _, item := range registry.Cases {
		if item.Path == path && item.ID == "explicit-consumer-continuation" &&
			item.Scope == ScopeLanguageCapability && item.Kind == KindValid && item.OrderedCollections {
			found++
		}
	}
	if found != 1 || !slices.Contains(registry.MetaSources, "internal/meta/executioncontinuation/policy/source_recovery.gooo") {
		t.Fatal("continuation source must occur once in executed corpus", found)
	}
	inventory, err := InspectSourceInventory(repository, raw)
	if err != nil || len(inventory.Unregistered) != 0 || len(inventory.Missing) != 0 {
		t.Fatal("source inventory incomplete", inventory.Unregistered, inventory.Missing, err)
	}
	result := replay.ExecuteWithOrderedCollections(repository, path, KindValid, "")
	if result.ObservedDecision != DecisionPass || !result.ASTReplayed || !result.ByteReplayed ||
		!result.SemanticReplayed || !result.GetPut || !result.PutGet {
		t.Fatalf("continuation source preservation laws: %+v", result)
	}
}

func TestContinuationSourceFamilyRetainsAllActivities(t *testing.T) {
	paths := []string{"internal/meta/executioncontinuation/policy/source.gooo",
		"internal/meta/executioncontinuation/policy/source_recovery.gooo"}
	sources, err := sourcebundle.Read(os.DirFS("../../../.."), paths)
	if err != nil {
		t.Fatal(err)
	}
	support := syntax.EntityFieldsV5Support()
	parsed, err := sourcebundle.ParseWithEntityFieldsSupport(sources, support)
	if err != nil {
		t.Fatal(err)
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(parsed.File, support)
	if err != nil {
		t.Fatal(err)
	}
	model, err := bidir.GetWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Nodes) != 12 {
		t.Fatal("three types and nine activities must remain", len(model.Nodes))
	}
	for _, role := range []string{"bind-subjects", "permit-continuation", "permit-wait", "bind-target-run",
		"observe-stage", "artifact-label", "workflow-path", "missing-dependency", "next-observation"} {
		id := "gooo://self-improvement/consumer/" + role
		found := false
		for _, node := range model.Nodes {
			if string(node.ID) == id && node.Kind == bidir.ActivityKind {
				found = true
			}
		}
		if !found {
			t.Fatal("source family lost activity", id)
		}
	}
	if err := bidir.CheckGetPutWithEntityFieldsSupport(document, support); err != nil {
		t.Fatal(err)
	}
	if err := bidir.CheckPutGetWithEntityFieldsSupport(document, model, support); err != nil {
		t.Fatal(err)
	}
}
