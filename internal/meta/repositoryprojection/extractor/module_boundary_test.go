package extractor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNestedModuleDependencyUsesOwningManifest(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOTOOLCHAIN", "local")
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module example.test/root\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "dep/go.mod", "module example.test/dep\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "dep/value.go", "package dep\nfunc Value() int { return 7 }\n")
	writeImporterFixture(t, root, "nested/go.mod", `module example.test/nested

go 1.27.3

require example.test/dep v0.0.0
replace example.test/dep => ../dep
`)
	writeImporterFixture(t, root, "nested/source.go", `package nested
import "example.test/dep"
func Value() int { return dep.Value() }
`)
	t.Chdir(root)
	result, err := ExtractWithResult(".", "nested/source.go")
	if err != nil {
		t.Fatalf("owning nested module dependency was rejected: %v", err)
	}
	if _, ok := result.Generated["nested/source.go"]; !ok {
		t.Fatalf("source was not retained in owning module: %v", result.Paths)
	}
	for _, path := range result.Paths {
		if !strings.HasPrefix(path, "nested/") {
			t.Fatalf("write set crossed owning module: %v", result.Paths)
		}
	}
}

func TestNestedModuleDoesNotBorrowRootDependency(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOTOOLCHAIN", "local")
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", `module example.test/root

go 1.27.3

require example.test/dep v0.0.0
replace example.test/dep => ./dep
`)
	writeImporterFixture(t, root, "dep/go.mod", "module example.test/dep\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "dep/value.go", "package dep\nfunc Value() int { return 7 }\n")
	writeImporterFixture(t, root, "nested/go.mod", "module example.test/nested\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "nested/source.go", `package nested
import "example.test/dep"
func Value() int { return dep.Value() }
`)
	_, err := ExtractWithResult(root, "nested/source.go")
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "UNRESOLVED_IMPORT" {
		t.Fatalf("root dependency leaked across module boundary: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "nested", "source_extracted.go")); !os.IsNotExist(err) {
		t.Fatalf("failed extraction wrote a helper: %v", err)
	}
}

func TestNestedModuleRejectsSourceSymlinkToRootModule(t *testing.T) {
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module example.test/root\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "root.go", "package root\nfunc Value() int { return 7 }\n")
	writeImporterFixture(t, root, "nested/go.mod", "module example.test/nested\n\ngo 1.27.3\n")
	if err := os.Symlink(filepath.Join("..", "root.go"), filepath.Join(root, "nested", "source.go")); err != nil {
		t.Fatal(err)
	}
	_, err := ExtractWithResult(root, "nested/source.go")
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "WRITE_SET_ESCAPE" {
		t.Fatalf("cross-module source link was accepted: %v", err)
	}
}

func TestNestedModuleRequiresCompatibleToolchainWithoutImports(t *testing.T) {
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module example.test/root\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "nested/go.mod", "module example.test/nested\n\ngo 1.99.0\n")
	writeImporterFixture(t, root, "nested/source.go", "package nested\nfunc Value() int { return 7 }\n")
	_, err := ExtractWithResult(root, "nested/source.go")
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "GO_MODULE_TOOLCHAIN_UNAVAILABLE" {
		t.Fatalf("newer module was rewritten with the local toolchain: %v", err)
	}
}
