package syntax

import (
	"strings"
	"testing"
)

func TestActivityExplicitIdentityParsesFormatsAndClones(t *testing.T) {
	source := "package p\nnamespace p\nactivity Calculate(Integer) -> Integer id \"urn:gooo:activity:calculate\" computes \"return input\"\n"
	file, diagnostics := ParseFile("identity.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	a := file.Declarations[0].(*ActivityDecl)
	if !a.IDPresent || a.ID != "urn:gooo:activity:calculate" || source[a.IDSpan.Start.Offset:a.IDSpan.End.Offset] != `"urn:gooo:activity:calculate"` {
		t.Fatal("explicit identity or its source span lost", a)
	}
	formatted, err := Format(file)
	if err != nil || !strings.Contains(formatted, `-> Integer id "urn:gooo:activity:calculate" computes "return input"`) {
		t.Fatal(formatted, err)
	}
	cloned := file.Clone().Declarations[0].(*ActivityDecl)
	cloned.ID = "urn:gooo:activity:other"
	if a.ID != "urn:gooo:activity:calculate" || !cloned.IDPresent || cloned.IDSpan != a.IDSpan {
		t.Fatal("clone changed source identity or dropped its span")
	}
}

func TestActivityIdentityRequiresQuotedValue(t *testing.T) {
	for _, suffix := range []string{"id bare", "id", `id "first" id "second"`} {
		_, diagnostics := ParseFile("bad.gooo", "package p\nnamespace p\nactivity A(Integer) -> Integer "+suffix)
		if !diagnostics.HasErrors() {
			t.Fatal("invalid activity identity accepted", suffix)
		}
	}
}
