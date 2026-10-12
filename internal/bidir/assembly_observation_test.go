package bidir

import (
	"strings"
	"testing"
)

func TestSourceObserveBXLawsAndSemanticIdentity(t *testing.T) {
	source := `package observed
namespace observed
entity Integer id "observed://integer"
activity Fee(Integer) -> Integer computes "return input - 11" assembling {
 choice "direction" operand_order at "0" intent "Subtract from eleven."
 case "11" -> "0"
 observe { input "17" oracle "Reference" max_candidates "2" max_rounds "1" }
 attempts "2"
}
activity Reference(Integer) -> Integer computes "return 11 - input"`
	doc := assemblyDocument(t, []byte(source))
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
		t.Fatal("observe Get-Put changed meaning", err)
	}
	core, err := LowerDocument(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range core.Graph.Nodes() {
		if node.Name != "Fee" {
			continue
		}
		if node.Assembly == nil || node.Assembly.Observation == nil || node.Assembly.Observation.Inputs[0] != 17 {
			t.Fatal("semantic IR lost source observation")
		}
		node.Assembly.Observation.Inputs[0] = 99
		fresh, ok := core.Graph.Node(node.ID)
		if !ok || fresh.Assembly.Observation.Inputs[0] != 17 {
			t.Fatal("graph exposed mutable observation storage")
		}
	}
	for _, change := range []struct{ from, to string }{
		{`input "17"`, `input "19"`},
		{`oracle "Reference"`, `oracle "Other"`},
		{`max_rounds "1"`, `max_rounds "2"`},
	} {
		changed := assemblyDocument(t, []byte(strings.Replace(source, change.from, change.to, 1)))
		changedModel, err := Get(changed)
		if err != nil {
			t.Fatal(err)
		}
		changedCore, err := LowerDocument(changed)
		if err != nil || changedCore.StableHash() == core.StableHash() ||
			SemanticFingerprint(changedModel) == SemanticFingerprint(model) {
			t.Fatal("observation change did not alter source meaning", change, err)
		}
	}
}
