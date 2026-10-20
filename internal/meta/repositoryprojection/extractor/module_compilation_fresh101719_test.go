package extractor

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func compilationFixture101719(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module experiment.compilation101719\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "dep/value.go", `package dep
import "context"
type Number int
func Value(ctx context.Context) Number { if ctx.Err()!=nil{return 0};return 41 }
`)
	writeImporterFixture(t, root, "consumer.go", `package consumer
import "context"
import "experiment.compilation101719/dep"
func Use(ctx context.Context) dep.Number { return dep.Value(ctx) }
`)
	return root
}
func TestCompiledBodyOriginsAndFreshHeader101719(t *testing.T) {
	root := compilationFixture101719(t)
	session := newStandardImportSession()
	first := newModuleImporterWithStandard(root, session)
	start := time.Now()
	pkg, err := first.Import("experiment.compilation101719/dep")
	if err != nil {
		t.Fatal(err)
	}
	contextBefore, err := first.Import("context")
	if err != nil {
		t.Fatal(err)
	}
	numberBefore := pkg.Scope().Lookup("Number").Type()
	function := pkg.Scope().Lookup("Value").Type().(*types.Signature)
	if numberBefore.Underlying().(*types.Basic).Kind() != types.Int || !types.Identical(function.Params().At(0).Type(), contextBefore.Scope().Lookup("Context").Type()) {
		t.Fatal("standard identity or declared type changed")
	}
	originBefore := first.compilationOrigins[pkg.Path()]
	if originBefore.BuildID == "" || originBefore.ExportSHA256 == "" || originBefore.SourceSHA256 == "" || originBefore.ContextSHA256 == "" || originBefore.SourceFiles != 1 || first.listCalls != 2 {
		t.Fatal("compilation origin absent", originBefore, first.listCalls)
	}
	writeImporterFixture(t, root, "dep/value.go", `package dep
import "context"
type Number float64
func Value(ctx context.Context) Number { if ctx.Err()!=nil{return 0};return 41.75 }
`)
	second := newModuleImporterWithStandard(root, session)
	next, err := second.Import(pkg.Path())
	if err != nil {
		t.Fatal(err)
	}
	contextAfter, err := second.Import("context")
	if err != nil {
		t.Fatal(err)
	}
	originAfter := second.compilationOrigins[pkg.Path()]
	if next == pkg || next.Scope().Lookup("Number").Type().Underlying().(*types.Basic).Kind() != types.Float64 || contextBefore != contextAfter || originAfter.SourceSHA256 == originBefore.SourceSHA256 || originAfter.BuildID == originBefore.BuildID || originAfter.ExportSHA256 == originBefore.ExportSHA256 {
		t.Fatal("source revision or standard universe lost", originBefore, originAfter)
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "consumer.go", `package consumer
import "context"
import "experiment.compilation101719/dep"
func Use(ctx context.Context) dep.Number { return dep.Value(ctx) }
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = (&types.Config{Importer: second}).Check("consumer101719", set, []*ast.File{file}, nil); err != nil {
		t.Fatal("fresh analyzed body", err)
	}
	summary := map[string]any{"first": originBefore, "next": originAfter, "first_list_calls": first.listCalls, "next_list_calls": second.listCalls, "source_type_transition": "int -> float64", "shared_standard_identity": true, "wall_ns": time.Since(start).Nanoseconds(), "historical_cohort_reruns": 0}
	if directory := os.Getenv("GOOO_COMPILATION_SOURCE_OUTPUT"); directory != "" {
		raw, _ := json.MarshalIndent(summary, "", "  ")
		if err = os.WriteFile(filepath.Join(directory, "compiled101719.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("new compiled input bodies; two metadata calls per stage; fresh source types; wall=%s", time.Since(start))
}
func TestInvalidDependencyBodyIsUnresolved101719(t *testing.T) {
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module invalid.compilation101719\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "dep/value.go", "package dep\nfunc Value() int {return \"invalid-body101719\"}\n")
	importer := newModuleImporter(root)
	if pkg, err := importer.Import("invalid.compilation101719/dep"); err == nil || pkg != nil || len(importer.compilationOrigins) != 0 {
		t.Fatal("uncompiled body supplied header types", pkg, err)
	}
}
