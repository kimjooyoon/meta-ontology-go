package extractor

import (
	"go/types"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompilationCaptureEmbeddedOrigin144617(t *testing.T) {
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module embed144617.example/scope\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "entry.go", "package capture144617\nimport _ \"embed\"\n//go:embed assets/note.txt\nvar Payload string\nfunc Value() string{return Payload}\n")
	writeImporterFixture(t, root, "assets/note.txt", "first source144617\n")
	imports := newModuleImporter(root)
	pkg, err := imports.Import("embed144617.example/scope")
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Scope().Lookup("Value").Type().(*types.Signature).Results().At(0).Type().String() != "string" || imports.compilationSourceFiles != 2 || imports.listCalls != 2 {
		t.Fatal(pkg, imports.compilationSourceFiles, imports.listCalls)
	}
	input := imports.compilationInputs[pkg.Path()]
	seen := false
	for _, source := range input.Inputs {
		if filepath.Base(source.Name) == "note.txt" {
			seen = !source.Go && string(source.Bytes) == "first source144617\n"
		}
	}
	if !seen {
		t.Fatal("embedded source origin absent", input.Inputs)
	}
	first := imports.compilationOrigins[pkg.Path()]
	writeImporterFixture(t, root, "assets/note.txt", "next runtime content144619\n")
	if _, err = imports.captureCompilationSources([]moduleCompiledPackage{input.Package}); err == nil {
		t.Fatal("changed embedded source joined the old type stage")
	}
	next := newModuleImporter(root)
	current, err := next.Import(pkg.Path())
	if err != nil {
		t.Fatal(err)
	}
	origin := next.compilationOrigins[pkg.Path()]
	if current == pkg || origin.SourceSHA256 == first.SourceSHA256 || origin.BuildID == first.BuildID {
		t.Fatal("new embedded input did not obtain a fresh origin", origin, first)
	}
	if !strings.Contains(origin.ImportPath, "embed144617") {
		t.Fatal(origin)
	}
}

func TestRefactoredCurrentDeclarationCapacity144617(t *testing.T) {
	targets := []string{"cmd/gooo#runBodyOrigins", "cmd/gooo#writeTypedBodyOrigins", "cmd/gooo#parseBodyOriginOptions", "cmd/gooo#validateBodyOriginOptions",
		"internal/meta/repositoryprojection/extractor#moduleImporter.captureCompilationSources", "internal/meta/repositoryprojection/extractor#compilationCapture.packageInput",
		"internal/meta/repositoryprojection/extractor#compilationCapture.directory", "internal/meta/repositoryprojection/extractor#compilationCapture.reserveFiles", "internal/meta/repositoryprojection/extractor#compilationCapture.sourceInput"}
	requireNamedCapacities(t, "../../../..", targets)
}
