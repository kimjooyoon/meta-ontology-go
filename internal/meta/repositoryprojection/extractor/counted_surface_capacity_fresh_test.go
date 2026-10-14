package extractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestCounterSurfaceCapacity19823(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	paths := []string{"internal/bodycodegen/body_surface.go", "internal/bodycodegen/counted_loop.go", "internal/bodycodegen/counted_loop_eval.go", "internal/bodycodegen/counted_loop_bindings.go", "internal/bodycodegen/counted_loop_inference.go", "internal/bodyexecution/composition_local_projection_render.go", "internal/packageruntime/workspaceexecution/local_file_ranges.go"}
	methods := 0
	for _, path := range paths {
		source, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, path, source, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			observed := observeRenderedCapacity(set, file, source, fn)
			if observed.functionStatus != renderedCapacityWithinCap || observed.helperStatus != renderedCapacityWithinCap || observed.helperLines == nil {
				t.Fatalf("%s:%s rendered capacity: %+v", path, fn.Name.Name, observed)
			}
			methods++
			t.Log(path, fn.Name.Name, "rendered", *observed.helperLines)
		}
	}
	if methods < 20 {
		t.Fatal("source inspection population", methods)
	}
	t.Log("new scoped-counter source functions inspected;unchanged75-line rendered bound;no native source or model execution")
}
