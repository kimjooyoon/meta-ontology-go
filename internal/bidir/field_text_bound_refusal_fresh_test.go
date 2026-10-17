package bidir

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestSourceBoundTypeAndSpellingRefusalsFresh1905(t *testing.T) {
	const original = `package bounds
namespace bounds
entity Box id "bound1905://box" fields {
 field value id "bound1905://value" type string required one max_utf8_bytes "1030"
}
`
	for _, tc := range []struct {
		name, from, to string
		parseError     bool
	}{
		{"boolean", "type string", "type boolean", false},
		{"integer", "type string", "type integer", false},
		{"negative", `"1030"`, `"-1"`, true},
		{"overflow", `"1030"`, `"4294967296"`, true},
		{"missing-value", `max_utf8_bytes "1030"`, `max_utf8_bytes`, true},
		{"duplicate", `max_utf8_bytes "1030"`, `max_utf8_bytes "1030" max_utf8_bytes "1031"`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := strings.Replace(original, tc.from, tc.to, 1)
			file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("bound.gooo", source, syntax.EntityFieldsV6Support())
			if tc.parseError {
				if !diagnostics.HasErrors() {
					t.Fatal("invalid source declaration parsed", source)
				}
				return
			}
			if diagnostics.HasErrors() {
				t.Fatal(diagnostics)
			}
			document, err := DocumentFromSyntaxWithEntityFieldsSupport(file, EntityFieldsV6Support())
			if err == nil {
				_, err = GetWithEntityFieldsSupport(document, EntityFieldsV6Support())
			}
			if err == nil || !strings.Contains(err.Error(), "max_utf8_bytes") {
				t.Fatal("non-Text field accepted a Text domain", err)
			}
		})
	}
}
