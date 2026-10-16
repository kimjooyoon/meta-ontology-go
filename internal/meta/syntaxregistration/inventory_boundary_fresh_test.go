package syntaxregistration

import (
	"go/ast"
	"strings"
	"testing"
)

const inventoryImport1892 = "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"

func inventorySource1892(t *testing.T, body, imported string) *goSource {
	t.Helper()
	path := syntaxRoot + "conformance/relocated_boundary1892_test.go"
	raw := "package languagesyntax_test\nimport languagesyntax \"" + imported + "\"\nfunc Boundary1892(){" + body + "}\n"
	source, err := parseSourceUnits(map[string][]byte{path: []byte(raw)}, syntaxRoot+"conformance/", "languagesyntax_test", true)
	if err != nil {
		t.Fatal(err)
	}
	source.activity = "gooo://syntax-registration/test-inventory-boundary"
	return source
}

func TestDeclaredInventoryBoundaryFresh1892(t *testing.T) {
	for _, row := range []struct {
		name, rhs, imported, edit string
		promote, accept           bool
	}{
		{"declared-append", "languagesyntax.DeclaredSourceCount()", inventoryImport1892, "", false, true},
		{"declared-promotion", "languagesyntax.DeclaredSourceCount()", inventoryImport1892, "", true, true},
		{"legacy-append", "150", inventoryImport1892, "151", false, true},
		{"legacy-promotion", "150", inventoryImport1892, "150", true, true},
		{"zero", "0", inventoryImport1892, "", false, false},
		{"negative", "-1", inventoryImport1892, "", false, false},
		{"wrong-import", "languagesyntax.DeclaredSourceCount()", "example.invalid/not-the-kernel", "", false, false},
		{"observation-argument", "languagesyntax.DeclaredSourceCount(report.Source.GoooFiles)", inventoryImport1892, "", false, false},
		{"observed-denominator", "len(report.Source.GoooFiles)", inventoryImport1892, "", false, false},
		{"unknown-method", "languagesyntax.ObservedSourceCount()", inventoryImport1892, "", false, false},
		{"missing-call", "languagesyntax.DeclaredSourceCount", inventoryImport1892, "", false, false},
		{"unknown-local", "sourceCount", inventoryImport1892, "", false, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			source := inventorySource1892(t, "if len(report.Source.GoooFiles) != "+row.rhs+" { panic(\"inventory\") }", row.imported)
			var comparison *ast.BinaryExpr
			ast.Inspect(source.file, func(node ast.Node) bool {
				if candidate, ok := node.(*ast.BinaryExpr); ok && comparison == nil {
					comparison = candidate
				}
				return true
			})
			err := reviseInventoryComparison(source, comparison, row.promote)
			if (err == nil) != row.accept {
				t.Fatal("inventory boundary selection differs", err)
			}
			if row.accept && row.edit == "" && len(source.edits) != 0 {
				t.Fatal("declaration-based count acquired a physical count edit")
			}
			if row.edit != "" && (len(source.edits) != 1 || source.edits[0].text != row.edit) {
				t.Fatal("historical literal transformation changed", source.edits)
			}
		})
	}
}

func TestInventoryFailureCannotBeErasedFresh1893(t *testing.T) {
	const semantic = `if report.Summary.Satisfied != 1 || report.Summary.ValidCases != 1 ||
report.Summary.Unresolved != 1 || languagesyntax.FixedTotal != 1 ||
languagesyntax.FixedCapabilityTotal != 1 { panic("semantic") }
`
	const raw = `{"cases":[{"kind":"VALID","scope":"LANGUAGE_CAPABILITY"}]}`
	for _, mode := range []string{"declared", "invalid-before-valid", "missing-semantic"} {
		t.Run(mode, func(t *testing.T) {
			body := semantic + `if len(report.Source.GoooFiles) != languagesyntax.DeclaredSourceCount() { panic("inventory") }`
			if mode == "invalid-before-valid" {
				body = `if len(report.Source.GoooFiles) != 0 { panic("invalid") };` + body
			}
			if mode == "missing-semantic" {
				body = strings.Replace(body, "report.Summary.Unresolved != 1 ||", "", 1)
			}
			source := inventorySource1892(t, body, inventoryImport1892)
			err := generateSyntaxTests(source, []byte(raw), false)
			if (err == nil) != (mode == "declared") {
				t.Fatal("inventory recognition weakened another obligation", mode, err)
			}
		})
	}
}
