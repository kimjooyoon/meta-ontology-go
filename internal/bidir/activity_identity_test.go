package bidir

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestLowerExplicitActivityIdentitySurvivesRename(t *testing.T) {
	for _, name := range []string{"Calculate", "합계"} {
		file, diagnostics := syntax.ParseFile("identity.gooo", `package p
namespace p
entity Integer id "urn:gooo:type:integer"
activity `+name+`(Integer) -> Integer id "urn:gooo:activity:sum" computes "return input"`)
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		ir, err := Lower(file)
		if err != nil {
			t.Fatal(err)
		}
		if !ir.Graph.HasFact(semantic.FactKey{Subject: "urn:gooo:activity:sum", Predicate: semantic.Used, Object: "urn:gooo:type:integer"}) ||
			!ir.Graph.HasFact(semantic.FactKey{Subject: "urn:gooo:type:integer", Predicate: semantic.WasGeneratedBy, Object: "urn:gooo:activity:sum"}) {
			t.Fatal("renamed activity lost stable PROV identity", ir.SemanticCanonical())
		}
		document, err := DocumentFromSyntax(file)
		if err != nil || document.Declarations[1].ID != "urn:gooo:activity:sum" {
			t.Fatal("syntax adapter lost source identity", document, err)
		}
		if err := CheckGetPut(document); err != nil {
			t.Fatal("Get-Put", err)
		}
		model, err := Get(document)
		if err != nil {
			t.Fatal(err)
		}
		if err := CheckPutGet(document, model); err != nil {
			t.Fatal("Put-Get", err)
		}
		for i := range model.Nodes {
			if model.Nodes[i].ID == "urn:gooo:activity:sum" {
				model.Nodes[i].Name = "Updated"
			}
		}
		if err := CheckPutGet(document, model); err != nil {
			t.Fatal("renamed Put-Get", err)
		}
		written, err := Put(document, model)
		if err != nil || written.Declarations[1].Name != "Updated" || written.Declarations[1].ID != "urn:gooo:activity:sum" {
			t.Fatal("write-back rename changed activity identity", written, err)
		}
	}
}

func TestLowerRejectsEmptyInvalidAndDuplicateExplicitActivityIdentity(t *testing.T) {
	for _, tail := range []string{`id ""`, `id "relative"`, `id "urn:gooo:type:integer"`, `id "urn:gooo:activity:sum"
activity Other(Integer) -> Integer id "urn:gooo:activity:sum"`} {
		file, diagnostics := syntax.ParseFile("bad.gooo", "package p\nnamespace p\nentity Integer id \"urn:gooo:type:integer\"\nactivity Sum(Integer) -> Integer "+tail)
		if diagnostics.HasErrors() {
			t.Fatal("test requires syntactically valid identity", diagnostics)
		}
		if _, err := Lower(file); err == nil {
			t.Fatal("invalid explicit identity silently changed or accepted", strings.ReplaceAll(tail, "\n", " "))
		}
	}
}
