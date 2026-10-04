package bidir

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func inputSequenceDocument(t *testing.T, source string) Document {
	t.Helper()
	file, diagnostics := syntax.ParseFile("joins.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestIndexedRepeatedInputsSurviveGetPutAndSemanticLowering(t *testing.T) {
	raw, err := os.ReadFile("../../examples/body-codegen/native-input-joins.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	document := inputSequenceDocument(t, string(raw))
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	var add Node
	for _, node := range model.Nodes {
		if node.Name == "Add" {
			add = node
		}
	}
	if len(add.InputSequence) != 2 || add.InputSequence[0].ID != "joins://integer" || add.InputSequence[1].ID != "joins://integer" || len(modelRuntimePorts(model, add.ID, PredicateUsed, true)) != 1 {
		t.Fatalf("source slots lost to PROV deduplication: %+v", add)
	}
	written, err := Put(document, model)
	if err != nil || !DocumentEquivalent(document, written) {
		t.Fatalf("Get-Put changed source: %v", err)
	}
	observed, err := Get(written)
	if err != nil || !SemanticEquivalent(model, observed) {
		t.Fatalf("Get-Put changed meaning: %v", err)
	}
	ir, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.Validate(); err != nil {
		t.Fatal(err)
	}
	node, _ := ir.Graph.Node(semantic.ID(add.ID))
	if !reflect.DeepEqual(node.InputSequence, []semantic.ID{"joins://integer", "joins://integer"}) {
		t.Fatal("IR lost repeated positions")
	}
	clone := ir.Graph.Clone()
	copy, _ := clone.Node(node.ID)
	copy.InputSequence[0] = "joins://text"
	original, _ := ir.Graph.Node(node.ID)
	if original.InputSequence[0] != "joins://integer" {
		t.Fatal("graph input sequence aliases a returned node")
	}
	if node.StableHash() == copy.StableHash() {
		t.Fatal("input type order is absent from node identity")
	}
	tampered := model.Clone()
	for i := range tampered.Nodes {
		if tampered.Nodes[i].ID == add.ID {
			tampered.Nodes[i].InputSequence[0].ID = "joins://text"
		}
	}
	if err := tampered.Validate(); err == nil {
		t.Fatal("input sequence disagreed with used facts")
	}
	if SemanticFingerprint(tampered) == SemanticFingerprint(model) {
		t.Fatal("sequence mutation did not change model fingerprint")
	}
}

func TestIndexedPortErrorsRemainExplicitAcrossAllPlans(t *testing.T) {
	source := `package joins
namespace joins
entity Integer id "joins://integer"
activity A(Integer) -> Integer computes "return input"
activity B(Integer, Integer) -> Integer computes "return input0 - input1"
bind A.result -> B.input0
`
	for _, port := range []string{"input", "input01", "input2"} {
		document := inputSequenceDocument(t, strings.Replace(source, "B.input0", "B."+port, 1))
		if _, err := Get(document); err == nil {
			t.Fatalf("Get accepted %q", port)
		}
		if _, err := LowerDocument(document); err == nil {
			t.Fatalf("IR accepted %q", port)
		}
		if _, err := CompileTypedPlan(document); err == nil {
			t.Fatalf("typed plan accepted %q", port)
		}
	}
	for _, suffix := range []string{"bind A.result -> B.input0", "bind B.result -> A.input"} {
		document := inputSequenceDocument(t, source+"\n"+suffix+"\n")
		if _, err := Get(document); err == nil {
			t.Fatalf("invalid incoming/cyclic edge accepted: %s", suffix)
		}
	}
}

func TestOrderedPrimitiveRootInputsAffectMeaningAndSurviveRename(t *testing.T) {
	document := inputSequenceDocument(t, `package roots
namespace roots
entity Boolean id "roots://boolean"
entity Text id "roots://text"
activity Pick(Boolean, Text) -> Text computes "if input0 { return input1 } else { return input1 }"
`)
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	reversed := model.Clone()
	for i := range reversed.Nodes {
		if reversed.Nodes[i].Kind == ActivityKind {
			s := reversed.Nodes[i].InputSequence
			s[0], s[1] = s[1], s[0]
		}
	}
	if SemanticEquivalent(model, reversed) || SemanticFingerprint(model) == SemanticFingerprint(reversed) {
		t.Fatal("input order treated as presentation")
	}
	written, err := Put(document, reversed)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := Get(written)
	if err != nil || !SemanticEquivalent(reversed, observed) {
		t.Fatalf("Put-Get lost ordered inputs: %v", err)
	}
	for i := range model.Nodes {
		if model.Nodes[i].Name == "Text" {
			model.Nodes[i].Name = "Words"
		}
	}
	written, err = Put(document, model)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range written.Declarations {
		if d.Kind == ActivityKind && (d.Inputs[1].Name != "Words" || d.Inputs[1].ID == "") {
			t.Fatal("ordered reference lost renamed display/identity")
		}
	}
}
