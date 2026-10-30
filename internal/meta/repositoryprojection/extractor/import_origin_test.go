package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestUnresolvedImportRetainsOwningModuleAndObservationCause(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOTOOLCHAIN", "local")
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module example.test/root\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "nested/go.mod", "module example.test/nested\n\ngo 1.27.3\n")
	source := []byte("package nested\nimport \"example.test/dependency\"\nfunc Value() int { return dependency.Value() }\n")
	writeImporterFixture(t, root, "nested/source.go", string(source))
	err := ValidateSourceImports(root, "nested/source.go", source)
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "UNRESOLVED_IMPORT" || failure.UnknownClass != "DIRECT_MISSING" {
		t.Fatalf("missing import changed admission or source classification: %v", err)
	}
	if !slices.Equal(failure.BlockedBy, []string{"example.test/dependency"}) ||
		!slices.Contains(failure.Diagnostics, "module=example.test/nested") ||
		!slices.Contains(failure.Diagnostics, "logical=nested/source.go") ||
		!strings.Contains(strings.Join(failure.Diagnostics, "\n"), "GO_LIST_SUBPROCESS_FAILURE") ||
		!strings.Contains(strings.Join(failure.Diagnostics, "\n"), "import lookup disabled by -mod=readonly") {
		t.Fatalf("owning source or Go observation cause lost: %v / %v", failure, failure.Diagnostics)
	}
	restoreDeclaredImportFixture(t, root)
	if err := ValidateSourceImports(root, "nested/source.go", source); err != nil {
		t.Fatalf("restored owning module input was refused: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(root, "nested/source.go"))
	if err != nil || !bytes.Equal(source, after) {
		t.Fatal("import observation changed the source", err)
	}
}

func restoreDeclaredImportFixture(t *testing.T, root string) {
	t.Helper()
	writeImporterFixture(t, root, "dep/go.mod", "module example.test/dependency\n\ngo 1.27.3\n")
	writeImporterFixture(t, root, "dep/value.go", "package dependency\nfunc Value() int { return 7 }\n")
	writeImporterFixture(t, root, "nested/go.mod", `module example.test/nested

go 1.27.3

require example.test/dependency v0.0.0
replace example.test/dependency => ../dep
`)
}

func TestImportObservationRetainsTypedBudgetCauseWithoutInventingOne(t *testing.T) {
	for _, test := range []struct {
		name, want string
		cause      error
	}{
		{"budget", "GO_LIST_CALL_BUDGET", fmt.Errorf("outer context: %w", moduleFailure(moduleCauseCallBudget))},
		{"unclassified", "cause=module dependency could not be resolved safely", errors.New("unclassified observation")},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := importResolutionFailure(".", "nested/source.go", "example.test/nested", "example.test/dep", test.cause)
			var failure Failure
			if !errors.As(err, &failure) || failure.UnknownClass != "DIRECT_MISSING" ||
				failure.NextOperation != "restore-import-package" || len(failure.BlockedBy) != 1 {
				t.Fatalf("unobserved dependency changed admission: %+v", failure)
			}
			diagnostics := strings.Join(failure.Diagnostics, "\n")
			if !strings.Contains(diagnostics, test.want) || strings.Contains(diagnostics, "outer context") ||
				(test.name == "unclassified" && strings.Contains(diagnostics, "GO_LIST_")) {
				t.Fatalf("typed cause was flattened or invented: %v", failure.Diagnostics)
			}
		})
	}
}
