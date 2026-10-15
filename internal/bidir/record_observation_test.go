package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"strings"
	"testing"
)

func TestRecordObserveBXLawsAndIRIdentity(t *testing.T) {
	source := `package observed
namespace observed
entity Integer id "record-observe://integer"
entity Box id "record-observe://box" fields { field value id "record-observe://value" type integer required one }
activity Select(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "0 - input" intent "Negate."
 value_case "[0]" -> "{\"value\":0}"
 observe { value_input "[9007199254740993]" oracle "Expected" max_candidates "2" max_rounds "1" }
 attempts "2"
}
activity Expected(Integer) -> Box computes "return Box{value: 0 - input}"`
	support := EntityFieldsV4Support()
	document := func(source string) Document {
		file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("record.gooo", source, syntax.EntityFieldsV4Support())
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
	for _, node := range core.Graph.Nodes() {
		if node.Name != "Select" {
			continue
		}
		if node.Assembly == nil || node.Assembly.Observation == nil || node.Assembly.Observation.ValueInputs[0] != "[9007199254740993]" {
			t.Fatal(node)
		}
		node.Assembly.Observation.ValueInputs[0] = "[7]"
		fresh, _ := core.Graph.Node(node.ID)
		if fresh.Assembly.Observation.ValueInputs[0] != "[9007199254740993]" {
			t.Fatal("IR question aliases mutable storage")
		}
	}
	for _, change := range [][2]string{{"[9007199254740993]", "[7]"}, {`oracle "Expected"`, `oracle "Other"`}, {`max_rounds "1"`, `max_rounds "2"`}} {
		changed := document(strings.Replace(source, change[0], change[1], 1))
		next, err := GetWithEntityFieldsSupport(changed, support)
		if err != nil {
			t.Fatal(err)
		}
		lowered, err := LowerDocumentWithEntityFieldsSupport(changed, support)
		if err != nil || core.StableHash() == lowered.StableHash() || SemanticFingerprint(model) == SemanticFingerprint(next) {
			t.Fatal("record question edit lost meaning", err)
		}
		if err := CheckPutGetWithEntityFieldsSupport(doc, next, support); err != nil {
			t.Fatal(err)
		}
	}
}
