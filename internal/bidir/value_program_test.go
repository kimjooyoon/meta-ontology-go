package bidir

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestActivityValueProgramBindsAtBidirResolutionAndFailsClosedBelowIt(t *testing.T) {
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
	if _, err := LowerDocument(document); err == nil || !strings.Contains(err.Error(), "semantic IR does not support declaration attributes") {
		t.Fatalf("core IR silently accepted an unrepresentable value program: %v", err)
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
