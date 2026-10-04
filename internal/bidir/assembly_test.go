package bidir

import (
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func assemblyDocument(t *testing.T, source []byte) Document {
	t.Helper()
	file, diagnostics := syntax.Parse(string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestAssemblySurvivesBXLawsAndCoreSemanticIdentity(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	document := assemblyDocument(t, source)
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	roundtrip, err := Put(document, model)
	if err != nil {
		t.Fatal(err)
	}
	getPut, err := Get(roundtrip)
	if err != nil || !SemanticEquivalent(model, getPut) {
		t.Fatal("assembly violated Get-Put", err)
	}
	core, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct{ from, to string }{
		{`case "5" -> "24"`, `case "5" -> "25"`},
		{`attempts "8"`, `attempts "7"`},
		{`alternative "second"`, `alternative "first"`},
		{`두 번째 지역값`, `첫 번째 지역값`},
	} {
		changed := assemblyDocument(t, []byte(strings.Replace(string(source), mutation.from, mutation.to, 1)))
		changedModel, err := Get(changed)
		if err != nil {
			t.Fatal(err)
		}
		changedCore, err := LowerDocument(changed)
		if err != nil || core.StableHash() == changedCore.StableHash() ||
			SemanticFingerprint(model) == SemanticFingerprint(changedModel) {
			t.Fatal("assembly contract edit lost semantic identity", mutation, err)
		}
		putGet, err := Put(document, changedModel)
		if err != nil {
			t.Fatal(err)
		}
		restored, err := Get(putGet)
		if err != nil || !SemanticEquivalent(changedModel, restored) {
			t.Fatal("assembly violated Put-Get", err)
		}
	}
	nodes := core.Graph.Nodes()
	for _, node := range nodes {
		if node.Name == "Qualified" {
			if node.Assembly == nil || len(node.Assembly.Choices) != 3 || len(node.Assembly.Cases) != 5 {
				t.Fatal("core IR dropped the typed assembly", node)
			}
			node.Assembly.Choices[0].Intent = "changed"
			fresh, _ := core.Graph.Node(node.ID)
			if fresh.Assembly.Choices[0].Intent == "changed" {
				t.Fatal("graph lookup exposes mutable assembly storage")
			}
		}
	}
}
