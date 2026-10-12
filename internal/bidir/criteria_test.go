package bidir

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestCriteriaBXLawsAndIRBindings(t *testing.T) {
	source := `package criteria
namespace criteria
entity Integer id "criteria://integer"
entity Boolean id "criteria://boolean"
entity Box id "criteria://box" fields { field value id "criteria://value" type integer required one }
activity Select(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "0-input" intent "Negate."
 criteria { predicate "negative" activity "Negative" input "[9007199254740993]" holdout_input "[9]" }
 attempts "2"
}
activity Negative(Integer,Box) -> Boolean computes "return input1.value == 0-input0"`
	support := EntityFieldsV4Support()
	document := func(source string) Document {
		file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("criteria.gooo", source, syntax.EntityFieldsV4Support())
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		doc, err := DocumentFromSyntaxWithEntityFieldsSupport(file, support)
		if err != nil {
			t.Fatal(err)
		}
		return doc
	}
	doc := document(source)
	if err := CheckGetPutWithEntityFieldsSupport(doc, support); err != nil {
		t.Fatal(err)
	}
	model, err := GetWithEntityFieldsSupport(doc, support)
	if err != nil {
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
		if node.Assembly == nil || node.Assembly.Criteria == nil || node.Assembly.Criteria.Predicates[0].Activity != "Negative" ||
			node.Assembly.Criteria.Inputs[0] != "[9007199254740993]" {
			t.Fatal("IR lost criteria binding", node)
		}
		node.Assembly.Criteria.Inputs[0] = "[7]"
		fresh, _ := core.Graph.Node(node.ID)
		if fresh.Assembly.Criteria.Inputs[0] != "[9007199254740993]" {
			t.Fatal("IR criteria aliases mutable storage")
		}
	}
	if !found {
		t.Fatal("assembly activity missing")
	}
	for _, change := range [][2]string{{"[9007199254740993]", "[7]"}, {`activity "Negative"`, `activity "Other"`},
		{`predicate "negative"`, `predicate "amount"`}, {`holdout_input "[9]"`, `holdout_input "[11]"`}} {
		changed := document(strings.Replace(source, change[0], change[1], 1))
		next, err := GetWithEntityFieldsSupport(changed, support)
		if err != nil {
			t.Fatal(err)
		}
		lowered, err := LowerDocumentWithEntityFieldsSupport(changed, support)
		if err != nil || core.StableHash() == lowered.StableHash() || SemanticFingerprint(model) == SemanticFingerprint(next) {
			t.Fatal("criterion edit lost its source meaning", err)
		}
		if err := CheckPutGetWithEntityFieldsSupport(doc, next, support); err != nil {
			t.Fatal(err)
		}
	}
}
