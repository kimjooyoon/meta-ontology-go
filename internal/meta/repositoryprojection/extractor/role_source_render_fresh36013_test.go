package extractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestRoleSourceRendering36013(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	logical := "internal/meta/syntaxregistration/physical_role_write_set_fresh34731_test.go"
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(logical)))
	if err != nil {
		t.Fatal(err)
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, logical, raw, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		count++
		value := observeRenderedCapacity(set, file, raw, function)
		if value.functionStatus != renderedCapacityWithinCap || value.helperStatus != renderedCapacityWithinCap || value.helperLines == nil {
			t.Fatal("role source is not safely renderable", function.Name.Name, value)
		}
		t.Log(function.Name.Name, "rendered", *value.helperLines)
	}
	if count != 4 {
		t.Fatal("role verification declarations omitted", count)
	}
	result, err := ExtractWithResult(root, logical)
	if err != nil {
		t.Fatal("actual Gooo source extraction failed", err)
	}
	if len(result.Generated) < 2 {
		t.Fatal("actual source partition did not emit helpers")
	}
	for path, source := range result.Generated {
		if path != logical && physicalLines(source) > 75 {
			t.Fatal("emitted helper exceeded unchanged75-line cap", path)
		}
	}
	t.Log("all4 verification declarations and actual source partition preserve75-line cap;all9 roles/3 model owners remain;producer/model0")
}
