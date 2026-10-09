package formatter

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestSyntaxAdapterPreservesExplicitActivityIdentity(t *testing.T) {
	for _, id := range []string{"urn:gooo:activity:sum", "p://activity/calculate"} {
		source := "package p\nnamespace p\nentity Integer id \"urn:gooo:type:integer\"\n" +
			"activity Calculate(Integer) -> Integer id \"" + id + "\"\n"
		file, diagnostics := syntax.ParseFile("identity.gooo", source)
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		result := FormatAST(file, SyntaxAdapter{})
		if result.HasErrors() || !strings.Contains(result.Source, ` id "`+id+`"`) {
			t.Fatal("formatter dropped source identity", result)
		}
		next, diagnostics := syntax.ParseFile("formatted.gooo", result.Source)
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		before, err := bidir.Lower(file)
		if err != nil {
			t.Fatal(err)
		}
		after, err := bidir.Lower(next)
		if err != nil || before.SemanticCanonical() != after.SemanticCanonical() {
			t.Fatal("format changed semantic identity", err)
		}
		first, _ := (SyntaxAdapter{}).Adapt(file)
		second, _ := (SyntaxAdapter{}).Adapt(next)
		if first.SemanticFingerprint() != second.SemanticFingerprint() {
			t.Fatal("fingerprint changed")
		}
	}
}

func TestFormatterRejectsInvalidActivityIdentity(t *testing.T) {
	for _, id := range []string{"relative", "billing://entity/order"} {
		document := billingDocument()
		document.Declarations[2].ID = id
		result := Format(&document)
		if !result.HasErrors() || result.Source != "" {
			t.Fatal("invalid identity emitted", id, result)
		}
	}
	file, diagnostics := syntax.ParseFile("empty.gooo", "package p\nnamespace p\nentity Integer id \"urn:gooo:type:integer\"\nactivity Sum(Integer) -> Integer id \"\"")
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	result := FormatAST(file, SyntaxAdapter{})
	if !result.HasErrors() || result.Source != "" {
		t.Fatal("explicit empty identity became derived identity", result)
	}
}

func TestFormatterComparesCanonicalActivityIdentities(t *testing.T) {
	document := billingDocument()
	document.Declarations[0].ID = "BILLING://ENTITY/order"
	document.Declarations[2].ID = "billing://entity/order"
	if result := Format(&document); !result.HasErrors() || result.Source != "" {
		t.Fatal("URI casing hid activity/entity collision", result)
	}
	document.Declarations[2].ID = "URN:gooo:activity:sum"
	first := document.SemanticFingerprint()
	document.Declarations[0].ID = "billing://entity/order"
	document.Declarations[2].ID = "urn:gooo:activity:sum"
	if first == "" || first != document.SemanticFingerprint() {
		t.Fatal("canonical URI casing changed semantic fingerprint")
	}
}
