package bidir

import (
	"os"
	"strings"
	"testing"
)

func TestSourceConditionsSurviveBXLawsAndBindSemanticIdentity(t *testing.T) {
	raw, err := os.ReadFile("../../examples/body-codegen/source-condition-cases.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	doc := assemblyDocument(t, raw)
	model, err := Get(doc)
	if err != nil {
		t.Fatal(err)
	}
	roundtrip, err := Put(doc, model)
	if err != nil {
		t.Fatal(err)
	}
	again, err := Get(roundtrip)
	if err != nil || !SemanticEquivalent(model, again) {
		t.Fatal("condition Get-Put differs", err)
	}
	core, err := LowerDocument(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range core.Graph.Nodes() {
		if node.Name != "Choose" {
			continue
		}
		if node.Assembly == nil || len(node.Assembly.ConditionCases) != 3 ||
			node.Assembly.ConditionCases[0].Input != -9007199254740995 {
			t.Fatal("semantic lowering lost exact condition cases")
		}
		node.Assembly.ConditionCases[0].Expected = false
		fresh, ok := core.Graph.Node(node.ID)
		if !ok || !fresh.Assembly.ConditionCases[0].Expected {
			t.Fatal("graph exposed mutable condition cases")
		}
	}
	changed := assemblyDocument(t, []byte(strings.Replace(string(raw), `-> "true"`, `-> "false"`, 1)))
	changedModel, err := Get(changed)
	if err != nil {
		t.Fatal(err)
	}
	changedCore, err := LowerDocument(changed)
	if err != nil || changedCore.StableHash() == core.StableHash() ||
		SemanticFingerprint(changedModel) == SemanticFingerprint(model) {
		t.Fatal("intermediate expectation did not change semantic identity", err)
	}
}
