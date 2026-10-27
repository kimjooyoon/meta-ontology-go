package extractor

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompilationClosureAndPackageBudgetsRemainDistinct(t *testing.T) {
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module budget.example\n\ngo 1.27.3\n")
	var names []string
	var packages []moduleCompiledPackage
	for index := range 9 {
		name := fmt.Sprintf("input%d.go", index)
		writeImporterFixture(t, root, name, strings.Repeat(" ", moduleFileMax))
		names = append(names, name)
		packages = append(packages, moduleCompiledPackage{Dir: root,
			ImportPath: fmt.Sprintf("budget.example/part%d", index), GoFiles: []string{name}})
	}
	imports := newModuleImporter(root)
	inputs, err := imports.captureCompilationSources(packages)
	if err != nil || len(inputs) != 9 {
		t.Fatal("nine bounded packages must fit the compilation closure", len(inputs), err)
	}
	_, err = imports.captureCompilationSources([]moduleCompiledPackage{{Dir: root,
		ImportPath: "budget.example", GoFiles: names}})
	var failure *moduleDependencyFailure
	if !errors.As(err, &failure) || failure.code != moduleCauseSourceLimit {
		t.Fatal("a single package exceeded its unchanged 8 MiB limit", err)
	}
	imports.compilationSourceBytes = moduleStageSourceMax - moduleFileMax + 1
	_, err = imports.captureCompilationSources(packages[:1])
	if !errors.As(err, &failure) || failure.code != moduleCauseStageSourceLimit {
		t.Fatal("stage limit did not retain its own cause", err)
	}
	writeImporterFixture(t, root, "too-large.go", strings.Repeat(" ", moduleFileMax+1))
	if _, err := readBoundedModuleSource(filepath.Join(root, "too-large.go")); err == nil {
		t.Fatal("the unchanged 1 MiB single-file limit was bypassed")
	}
}
