package extractor

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCompilationRootsRetainFreshOriginsWithinSameCallBudget(t *testing.T) {
	root, source := independentImportFixture(t)
	file, err := parser.ParseFile(token.NewFileSet(), "consumer.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	imports := newModuleImporter(root)
	if err := imports.prepareCompilationImports([]*ast.File{file}); err != nil {
		t.Fatal(err)
	}
	if imports.listCalls != 2 || len(imports.compilationOrigins) != 34 || imports.compilationSourceFiles != 34 {
		t.Fatal("independent roots were not closed together", imports.listCalls, len(imports.compilationOrigins), imports.compilationSourceFiles)
	}
	writeImporterFixture(t, root, "dep33/value.go", "package dep33\nconst Value = \"changed\"\n")
	fresh := newModuleImporter(root)
	if err := fresh.prepareCompilationImports([]*ast.File{file}); err != nil {
		t.Fatal(err)
	}
	if fresh.compilationOrigins["independent.origins/dep33"].SourceSHA256 == imports.compilationOrigins["independent.origins/dep33"].SourceSHA256 {
		t.Fatal("new source reused the earlier stage's origin")
	}
	err = projectedFinalConformance(root, "consumer.go", map[string][]byte{"consumer.go": source})
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "PROJECTED_CONFORMANCE_FAILED" || failure.UnknownClass != "KNOWN_CONTRADICTION" {
		t.Fatal("a genuine changed dependency type was not refuted", err)
	}
}
