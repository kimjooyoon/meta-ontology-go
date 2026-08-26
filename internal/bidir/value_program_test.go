package bidir

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestActivityValueProgramBindsThroughCoreIRAndUnknownAttributesFailClosed(t *testing.T) {
	document := valueProgramDocument(t, "int.add:1")
	activity := document.Declarations[1]
	if got := activity.Attributes[ActivityValueProgramAttribute]; got != "int.add:1" {
		t.Fatalf("semantic value program = %q", got)
	}
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := Get(valueProgramDocument(t, "int.add:2"))
	if err != nil {
		t.Fatal(err)
	}
	if SemanticFingerprint(model) == SemanticFingerprint(changed) {
		t.Fatal("value program did not participate in the bidir semantic fingerprint")
	}
	core, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	var coreProgram string
	for _, node := range core.Graph.Nodes() {
		if node.Name == "Increment" {
			coreProgram = node.ValueProgram
		}
	}
	if coreProgram != "int.add:1" {
		t.Fatalf("core IR value program = %q", coreProgram)
	}
	changedCore, err := LowerDocument(valueProgramDocument(t, "int.add:2"))
	if err != nil {
		t.Fatal(err)
	}
	if core.StableHash() == changedCore.StableHash() {
		t.Fatal("value program did not participate in the core IR fingerprint")
	}
	unknown := document
	unknown.Declarations = append([]Declaration(nil), document.Declarations...)
	activity := unknown.Declarations[1]
	activity.Attributes = map[string]string{
		ActivityValueProgramAttribute: "int.add:1",
		"gooo:activity:unknown":       "must-fail-closed",
	}
	unknown.Declarations[1] = activity
	if _, err := LowerDocument(unknown); err == nil || !strings.Contains(err.Error(), "semantic IR does not support declaration attributes") {
		t.Fatalf("core IR accepted an unknown declaration attribute: %v", err)
	}
}

func valueProgramDocument(t *testing.T, program string) Document {
	t.Helper()
	source := "package valuewitness\nnamespace valuewitness\n" +
		"entity Integer id \"gooo://value-witness/entity/integer\"\n" +
		"activity Increment(Integer) -> Integer computes \"" + program + "\""
	file, diagnostics := syntax.Parse(source)
	if diagnostics.HasErrors() {
		t.Fatalf("parse diagnostics: %v", diagnostics)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	return document
}
