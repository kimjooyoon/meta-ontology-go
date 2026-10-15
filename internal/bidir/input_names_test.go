package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

func namedInputSource(t *testing.T, signature string) (string, *syntax.File) {
	t.Helper()
	source := "package named\nnamespace named\nentity Integer id \"urn:gooo:type:integer\"\n" +
		"# 입력 순서와 이름을 각각 보관한다.\nactivity Compute(" + signature + ") -> Integer id \"urn:named:compute\" computes `return 1`\n"
	file, diagnostics := syntax.ParseFile("named.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	return source, file
}

func TestAuthoredInputNamesSurviveSemanticViewsFirst(t *testing.T) {
	for _, signature := range []string{"amount Integer", "minuend Integer, subtrahend Integer", "Integer, adjustment Integer"} {
		source, file := namedInputSource(t, signature)
		document, err := DocumentFromSyntax(file)
		if err != nil {
			t.Fatal(err)
		}
		model, err := Get(document)
		if err != nil {
			t.Fatal(err)
		}
		if err := CheckGetPut(document); err != nil {
			t.Fatal("Get-Put names", err)
		}
		if err := CheckPutGet(document, model); err != nil {
			t.Fatal("Put-Get names", err)
		}
		ir, err := Lower(file)
		if err != nil {
			t.Fatal(err)
		}
		node, ok := ir.Graph.Node("urn:named:compute")
		if !ok {
			t.Fatal("stable activity ID missing")
		}
		checkSemanticInputNameSpans(t, source, node, document.Declarations[1].InputNames)
		changed := model.Clone()
		for i := range changed.Nodes {
			if changed.Nodes[i].ID == "urn:named:compute" {
				changed.Nodes[i].InputNames[0].Name = "replacement"
			}
		}
		if SemanticEquivalent(model, changed) || SemanticFingerprint(model) == SemanticFingerprint(changed) {
			t.Fatal("body binding names absent from semantic comparison")
		}
	}
}

func checkSemanticInputNameSpans(t *testing.T, source string, node semantic.Node, declared []Reference) {
	t.Helper()
	if len(node.InputNames) != len(declared) {
		t.Fatal("input names dropped during lowering")
	}
	for i, name := range node.InputNames {
		if name.Name != declared[i].Name {
			t.Fatal("input name order changed")
		}
		if name.Name != "" && source[name.Span.Start.Offset:name.Span.End.Offset] != name.Name {
			t.Fatal("input name detached from original UTF-8 source")
		}
	}
	if len(node.InputNames) == 2 && (len(node.InputSequence) != 2 || node.InputSequence[0] != "urn:gooo:type:integer" ||
		node.InputSequence[1] != "urn:gooo:type:integer") {
		t.Fatal("same-type positions collapsed")
	}
}
