package bidir

import (
	"strings"
	"testing"
)

func TestCriteriaStrategyChangesIRMeaningAndPreservesBXLaws(t *testing.T) {
	before := strings.Replace(criteriaBXSource, `attempts "2"`, `observe { criteria_input "[13]" max_candidates "2" max_rounds "1" } attempts "2"`, 1)
	after := strings.Replace(before, "observe {", `observe { criteria_strategy "distinguish_then_check"`, 1)
	support := EntityFieldsV4Support()
	doc, changed := criteriaDocument(t, before), criteriaDocument(t, after)
	if err := CheckGetPutWithEntityFieldsSupport(changed, support); err != nil {
		t.Fatal(err)
	}
	model, err := GetWithEntityFieldsSupport(changed, support)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPutGetWithEntityFieldsSupport(doc, model, support); err != nil {
		t.Fatal(err)
	}
	a, err := LowerDocumentWithEntityFieldsSupport(doc, support)
	if err != nil {
		t.Fatal(err)
	}
	b, err := LowerDocumentWithEntityFieldsSupport(changed, support)
	if err != nil || a.StableHash() == b.StableHash() {
		t.Fatal("strategy edit lost meaning", err)
	}
	found := false
	for _, node := range b.Graph.Nodes() {
		if node.Name == "Select" {
			found = node.Assembly.Observation.CriteriaStrategy == "distinguish_then_check"
		}
	}
	if !found {
		t.Fatal("IR strategy missing")
	}
}
