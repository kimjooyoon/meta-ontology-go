package extractor

import (
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompiledOriginPartitionInvariance134117(t *testing.T) {
	values := []string{}
	for i := range 257 {
		values = append(values, fmt.Sprintf("const Value%d = %d\n", i, 2441+i))
	}
	for _, split := range []bool{false, true} {
		root := t.TempDir()
		writeImporterFixture(t, root, "go.mod", "module partition134117.example/module\n\ngo 1.27.3\n")
		if split {
			for i, decl := range values {
				writeImporterFixture(t, root, fmt.Sprintf("part%03d.go", i), "package partition134117\n"+decl)
			}
		} else {
			writeImporterFixture(t, root, "complete.go", "package partition134117\n"+strings.Join(values, ""))
		}
		imports := newModuleImporter(root)
		pkg, err := imports.Import("partition134117.example/module")
		if err != nil {
			t.Fatal("compiled package partition changed admissibility", split, err)
		}
		object, ok := pkg.Scope().Lookup("Value256").(*types.Const)
		if !ok || object.Val().ExactString() != "2697" || len(pkg.Scope().Names()) != 257 {
			t.Fatal("public declarations differ", split, object)
		}
		for i := range 257 {
			constant, ok := pkg.Scope().Lookup(fmt.Sprintf("Value%d", i)).(*types.Const)
			if !ok || constant.Val().ExactString() != fmt.Sprint(2441+i) {
				t.Fatal("declared value changed with partition", split, i)
			}
		}
		want := 1
		if split {
			want = 257
		}
		origin := imports.compilationOrigins[pkg.Path()]
		if origin.SourceFiles != want || imports.compilationSourceFiles != want || len(imports.compilationInputs) != 1 || imports.listCalls != 2 {
			t.Fatal(origin, imports.compilationSourceFiles, imports.listCalls)
		}
		// An already retained exact input consumes its representation budget once.
		if _, err = imports.captureCompilationSources([]moduleCompiledPackage{imports.compilationInputs[pkg.Path()].Package}); err != nil || imports.compilationSourceFiles != want {
			t.Fatal(err, imports.compilationSourceFiles)
		}
	}
}

func TestCompilationStageFileBudget134117(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"first.go", "second.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("package partition134117\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	imports := newModuleImporter(root)
	imports.module = "partition134117.example/module"
	imports.compilationSourceFiles = moduleStageFilesMax - 1
	_, err := imports.captureCompilationSources([]moduleCompiledPackage{{Dir: root, ImportPath: imports.module, GoFiles: []string{"first.go", "second.go"}}})
	if failure, ok := err.(*moduleDependencyFailure); !ok || failure.code != moduleCauseFileLimit {
		t.Fatal("cumulative input budget lost", err)
	}
	if imports.compilationSourceFiles != moduleStageFilesMax-1 || len(imports.compilationInputs) != 0 {
		t.Fatal("failed capture retained a partial stage")
	}
}
