package bidir

import (
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func recordAssemblyDocument(t *testing.T, source string) Document {
	t.Helper()
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("record.gooo", source, syntax.EntityFieldsV1Support())
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := DocumentFromSyntaxWithEntityFieldsSupport(file, EntityFieldsV1Support())
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestRecordFieldAssemblyPreservesBXLawsAndSemanticContract(t *testing.T) {
	raw, err := os.ReadFile("../../examples/body-codegen/record-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	document := recordAssemblyDocument(t, source)
	support := EntityFieldsV1Support()
	model, err := GetWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	if err = CheckGetPutWithEntityFieldsSupport(document, support); err != nil {
		t.Fatal(err)
	}
	core, err := LowerDocumentWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"attempts \"8\"", "attempts \"7\""},
		{"Keep the original title.", "Use a new title."}, {"alternative \"input0.title\"", "alternative \"input0.reason\""},
		{"검토:accepted", "검토:changed"}} {
		changed := recordAssemblyDocument(t, strings.Replace(source, pair[0], pair[1], 1))
		next, err := GetWithEntityFieldsSupport(changed, support)
		if err != nil {
			t.Fatal(err)
		}
		lowered, err := LowerDocumentWithEntityFieldsSupport(changed, support)
		if err != nil || SemanticFingerprint(model) == SemanticFingerprint(next) || core.StableHash() == lowered.StableHash() {
			t.Fatal("record assembly edit lost semantic meaning", err)
		}
		if err = CheckPutGetWithEntityFieldsSupport(document, next, support); err != nil {
			t.Fatal(err)
		}
	}
	for _, node := range core.Graph.Nodes() {
		if node.Name == "Select" {
			if node.Assembly == nil || len(node.Assembly.ValueCases) != 5 {
				t.Fatal("record finite cases lost in core")
			}
			node.Assembly.ValueCases[0].Inputs = "changed"
			fresh, _ := core.Graph.Node(node.ID)
			if fresh.Assembly.ValueCases[0].Inputs == "changed" {
				t.Fatal("record cases share mutable storage")
			}
		}
	}
}
