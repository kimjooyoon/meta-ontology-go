package extractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestCountedReplayCapacity18829(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	source, err := os.ReadFile(filepath.Join(root, "internal/bodycodegen/source_fill_replay.go"))
	if err != nil {
		t.Fatal(err)
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "source_fill_replay.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || (fn.Name.Name != "compareSourceFillReplay" && fn.Name.Name != "compareSourceFillCompleteness") {
			continue
		}
		observed := observeRenderedCapacity(set, file, source, fn)
		if observed.functionStatus != renderedCapacityWithinCap || observed.helperStatus != renderedCapacityWithinCap || observed.helperLines == nil {
			t.Fatalf("new decomposition capacity: %+v", observed)
		}
		checked++
		t.Log(fn.Name.Name, "rendered helper lines", *observed.helperLines)
	}
	if checked != 2 {
		t.Fatal("decomposed replay functions", checked)
	}
	t.Log("two new source functions retain rendered capacity under unchanged75-line bound;source inspection only;producer/model0")
}
