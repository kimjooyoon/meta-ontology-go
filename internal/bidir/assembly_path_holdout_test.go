package bidir

import (
	"strings"
	"testing"
)

func TestTypedPathHoldoutBXLawsAndMeaning(t *testing.T) {
	source := `package withheld
namespace withheld
entity Integer id "withheld://integer"
activity Fee(Integer) -> Integer computes "return input - 19" assembling {
 choice "direction" operand_order at "0" intent "Choose a direction."
 case "19" -> "0"
 holdout_case "23" -> "-4"
 attempts "2"
}`
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
		t.Fatal("Get-Put changed held evaluation", err)
	}
	core, err := LowerDocument(doc)
	if err != nil {
		t.Fatal(err)
	}
	changedDoc := assemblyDocument(t, []byte(strings.Replace(source, `"23" -> "-4"`, `"23" -> "4"`, 1)))
	changed, err := Get(changedDoc)
	if err != nil {
		t.Fatal(err)
	}
	changedCore, err := LowerDocument(changedDoc)
	if err != nil {
		t.Fatal(err)
	}
	if core.StableHash() == changedCore.StableHash() || SemanticFingerprint(model) == SemanticFingerprint(changed) {
		t.Fatal("evaluation expectation failed to change source meaning")
	}
	for _, node := range core.Graph.Nodes() {
		if node.Name != "Fee" {
			continue
		}
		if node.Assembly == nil || len(node.Assembly.HoldoutCases) != 1 {
			t.Fatal("IR lost path holdout")
		}
		node.Assembly.HoldoutCases[0].Expected = 9
		fresh, ok := core.Graph.Node(node.ID)
		if !ok || fresh.Assembly.HoldoutCases[0].Expected != -4 {
			t.Fatal("IR exposed mutable evaluation")
		}
	}
}
