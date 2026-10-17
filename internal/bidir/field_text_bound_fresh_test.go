package bidir

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

func boundDocument1905(t *testing.T, suffix string) Document {
	t.Helper()
	source := `package bounded
namespace bounded
entity Expanded id "bound1905://expanded" fields {
 field value id "bound1905://expanded/value" type string required one` + suffix + `
}
`
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("bounded.gooo", source, syntax.EntityFieldsV6Support())
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := DocumentFromSyntaxWithEntityFieldsSupport(file, EntityFieldsV6Support())
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestTextBoundBXIdentityFresh1905(t *testing.T) {
	document := boundDocument1905(t, ` max_utf8_bytes "1030"`)
	support := EntityFieldsV6Support()
	model, err := GetWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	field := model.Nodes[0].Fields[0]
	if field.TextBound != textbound.New(1030) || field.ID != "bound1905://expanded/value" {
		t.Fatal("lowering lost source constraint or identity", field)
	}
	if err := CheckGetPutWithEntityFieldsSupport(document, support); err != nil {
		t.Fatal("source bound violated Get-Put", err)
	}
	if err := CheckPutGetWithEntityFieldsSupport(document, model, support); err != nil {
		t.Fatal("source bound violated Put-Get", err)
	}
	if _, err := GetWithEntityFieldsSupport(document, EntityFieldsV5Support()); err == nil {
		t.Fatal("older profile accepted source byte constraint")
	}
	if !strings.Contains(documentEvidence(document).Declarations[0].Fields[0].TextBound.Canonical(), "1030") {
		t.Fatal("source evidence lost the bound")
	}
}

func TestTextBoundSemanticDifferenceFresh1905(t *testing.T) {
	plain := boundDocument1905(t, "")
	zero := boundDocument1905(t, ` max_utf8_bytes "0"`)
	large := boundDocument1905(t, ` max_utf8_bytes "1030"`)
	models := make([]Model, 0, 3)
	for _, document := range []Document{plain, zero, large} {
		model, err := GetWithEntityFieldsSupport(document, EntityFieldsV6Support())
		if err != nil {
			t.Fatal(err)
		}
		models = append(models, model)
	}
	for _, model := range models {
		if err := CheckPutGetWithEntityFieldsSupport(plain, model, EntityFieldsV6Support()); err != nil {
			t.Fatal("writing a different declared domain lost field meaning", err)
		}
	}
	for i := range models {
		for j := i + 1; j < len(models); j++ {
			if SemanticFingerprint(models[i]) == SemanticFingerprint(models[j]) {
				t.Fatal("bound change did not change semantic fingerprint", i, j)
			}
			if fieldSemanticEqual(models[i].Nodes[0].Fields[0], models[j].Nodes[0].Fields[0]) {
				t.Fatal("undeclared, zero and expanded domains became identical", i, j)
			}
		}
	}
	renamed := models[2].Nodes[0].Fields[0]
	renamed.Name = "canonical"
	if !fieldSemanticEqual(models[2].Nodes[0].Fields[0], renamed) {
		t.Fatal("display rename changed field meaning")
	}
}
