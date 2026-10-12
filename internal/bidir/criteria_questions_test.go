package bidir

import (
	"strings"
	"testing"
)

func TestCriteriaQuestionsBXAndIRKeepExactInputs(t *testing.T) {
	source := strings.Replace(criteriaBXSource, `attempts "2"`, `observe { criteria_input "[-9007199254740993]" max_candidates "2" max_rounds "1" }
 attempts "2"`, 1)
	support := EntityFieldsV4Support()
	doc := criteriaDocument(t, source)
	if err := CheckGetPutWithEntityFieldsSupport(doc, support); err != nil {
		t.Fatal(err)
	}
	core, err := LowerDocumentWithEntityFieldsSupport(doc, support)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, node := range core.Graph.Nodes() {
		if node.Name != "Select" {
			continue
		}
		found = true
		if node.Assembly.Observation.CriteriaInputs[0] != "[-9007199254740993]" {
			t.Fatal(node)
		}
		node.Assembly.Observation.CriteriaInputs[0] = "[1]"
		fresh, _ := core.Graph.Node(node.ID)
		if fresh.Assembly.Observation.CriteriaInputs[0] != "[-9007199254740993]" {
			t.Fatal("IR aliases question storage")
		}
	}
	if !found {
		t.Fatal("assembly node missing")
	}
	changed := criteriaDocument(t, strings.Replace(source, "[-9007199254740993]", "[-7]", 1))
	model, err := GetWithEntityFieldsSupport(changed, support)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPutGetWithEntityFieldsSupport(doc, model, support); err != nil {
		t.Fatal(err)
	}
	next, err := LowerDocumentWithEntityFieldsSupport(changed, support)
	if err != nil || core.StableHash() == next.StableHash() {
		t.Fatal("question edit lost its identity", err)
	}
}
