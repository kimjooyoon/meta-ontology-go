package extractor

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestCurrentProjectDeclarationOrigins108271(t *testing.T) {
	_, name, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("source location unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(name), "../../../.."))
	source := `package fresh108271
import "github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
func NewSourceProbe108271() any { return generation.ObserveRegistrationInput }
`
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "source-consumer108271.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	imports := newModuleImporter(root)
	start := time.Now()
	pkg, err := (&types.Config{Importer: imports}).Check("fresh108271", set, []*ast.File{file}, nil)
	summary := map[string]any{"new_consumer_source": source, "wall_ns": time.Since(start).Nanoseconds(), "list_calls": imports.listCalls, "compiled_dependencies": len(imports.compilationOrigins), "old_test_executions": 0, "error": ""}
	if err != nil {
		summary["error"] = err.Error()
	}
	if output := os.Getenv("GOOO_COMPILATION_SOURCE_OUTPUT"); output != "" {
		raw, _ := json.MarshalIndent(summary, "", "  ")
		if writeErr := os.WriteFile(filepath.Join(output, "project108271.json"), raw, 0600); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err != nil || pkg == nil {
		t.Fatal("current project declared types", err)
	}
	if len(imports.compilationOrigins) == 0 || imports.listCalls > 3 {
		t.Fatal("source compilation work repeated", imports.listCalls, len(imports.compilationOrigins))
	}
	t.Logf("current project: new source consumer; compiled dependencies=%d; metadata calls=%d; elapsed=%s", len(imports.compilationOrigins), imports.listCalls, time.Since(start))
}
