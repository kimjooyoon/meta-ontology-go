package extractor

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestCompilationObservationRefusalDoesNotRefuteSource(t *testing.T) {
	root, source := independentImportFixture(t)
	file, err := parser.ParseFile(token.NewFileSet(), "consumer.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	imports := newModuleImporter(root)
	imports.listCalls = moduleListCallsMax
	err = imports.prepareCompilationImports([]*ast.File{file})
	var native *moduleDependencyFailure
	if !errors.As(err, &native) || native.code != moduleCauseCallBudget {
		t.Fatal("call budget was waived", err)
	}
	err = projectedDependencyOriginFailure(root, "logical=consumer.go", err)
	var failure Failure
	if !errors.As(err, &failure) || failure.UnknownClass != "DIRECT_MISSING" ||
		failure.Reason != "COMPILATION_ORIGIN_UNAVAILABLE" || !strings.Contains(strings.Join(failure.Diagnostics, "\n"), "GO_LIST_CALL_BUDGET") {
		t.Fatal("unobserved imports became a source contradiction or lost their cause", err)
	}
}

func TestUncompiledDependencyCannotSupplyProjectedTypes(t *testing.T) {
	root, source := independentImportFixture(t)
	writeImporterFixture(t, root, "dep33/value.go", "package dep33\nvar Value = missingInput\n")
	err := projectedFinalConformance(root, "consumer.go", map[string][]byte{"consumer.go": source})
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "COMPILATION_ORIGIN_UNAVAILABLE" ||
		failure.UnknownClass != "DIRECT_MISSING" || failure.NextOperation != "restore-module-compilation-origin" {
		t.Fatal("uncompiled dependency was admitted or source was falsely refuted", err)
	}
}
