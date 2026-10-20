package extractor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompilationInputBudgetIsStageOwned110319(t *testing.T) {
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module budget110319\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "dep/value.go", "package dep\nconst Value=57\n")
	directory := filepath.Join(root, "dep")
	source, err := os.ReadFile(filepath.Join(directory, "value.go"))
	if err != nil {
		t.Fatal(err)
	}
	importer := newModuleImporter(root)
	importer.compilationSourceBytes = moduleSourceMax - int64(len(source)) + 1
	observed := []moduleCompiledPackage{{Dir: directory, ImportPath: "budget110319/dep", GoFiles: []string{"value.go"}}}
	if _, err = importer.captureCompilationSources(observed); err == nil {
		t.Fatal("existing stage input population ignored")
	}
	importer.compilationSourceBytes = moduleSourceMax - int64(len(source))
	inputs, err := importer.captureCompilationSources(observed)
	if err != nil || len(inputs) != 1 {
		t.Fatal(inputs, err)
	}
	importer.retainCompilationOrigins(inputs, map[string]moduleCompilationOrigin{"budget110319/dep": {ImportPath: "budget110319/dep"}})
	if importer.compilationSourceBytes != moduleSourceMax {
		t.Fatal("retained input population lost", importer.compilationSourceBytes)
	}
	if _, err = importer.captureCompilationSources(observed); err != nil {
		t.Fatal("existing exact origin double-counted", err)
	}
	writeImporterFixture(t, root, "dep/value.go", "package dep\nconst Value=59\n")
	if importer.currentCompilationInputs(inputs) {
		t.Fatal("changed input retained old compilation origin")
	}
	if _, err = importer.captureCompilationSources(observed); err == nil {
		t.Fatal("two source versions combined in one type stage")
	}
}
