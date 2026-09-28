package extractor

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestProjectedConformancePreservesNativeFailurePhase(t *testing.T) {
	cases := []struct {
		name   string
		source string
		helper string
		phase  string
		cause  string
	}{
		{"target parse", "package p\nfunc F(\n", "", "parse-target", "expected"},
		{"helper parse", "package p\nfunc F() {}\n", "package p\nfunc H(\n", "parse-generated", "expected"},
		{"helper package", "package p\nfunc F() {}\n", "package other\nfunc H() {}\n", "package-name", "does not match target package"},
		{"target typecheck", "package p\nfunc F() int { return missingValue }\n", "", "typecheck", "missingValue"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n\ngo 1.27.0\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "x.go"), []byte("package p\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			generated := map[string][]byte{"x.go": []byte(tc.source)}
			if tc.helper != "" {
				generated["x_extracted.go"] = []byte(tc.helper)
			}
			failure := assertProjectedDiagnosticFailure(t, projectedFinalConformance(root, "x.go", generated), tc.phase, tc.cause)
			if strings.Contains(strings.Join(failure.Diagnostics, "\n"), root) {
				t.Fatalf("temporary workspace escaped into diagnostics: %#v", failure.Diagnostics)
			}
		})
	}
}

func TestProjectedConformanceDiagnosticRelocation(t *testing.T) {
	firstRoot, secondRoot := t.TempDir(), t.TempDir()
	first := projectedConformanceFailure(firstRoot, "logical=pkg/x.go", "typecheck",
		errors.New(filepath.Join(firstRoot, "pkg", "x.go")+":7: missingValue"))
	second := projectedConformanceFailure(secondRoot, "logical=pkg/x.go", "typecheck",
		errors.New(filepath.Join(secondRoot, "pkg", "x.go")+":7: missingValue"))
	left := assertProjectedDiagnosticFailure(t, first, "typecheck", "<workspace>/pkg/x.go:7: missingValue")
	right := assertProjectedDiagnosticFailure(t, second, "typecheck", "<workspace>/pkg/x.go:7: missingValue")
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("relocation changed native failure diagnostics: first=%#v second=%#v", left, right)
	}
}

func TestProjectedConformanceDiagnosticCauseIsBounded(t *testing.T) {
	err := projectedConformanceFailure(t.TempDir(), "logical=x.go", "typecheck", errors.New(strings.Repeat("x", 8192)))
	failure := assertProjectedDiagnosticFailure(t, err, "typecheck", "...[truncated]")
	cause := strings.TrimPrefix(failure.Diagnostics[2], "cause=")
	if len([]rune(cause)) != 4096+len("...[truncated]") {
		t.Fatalf("native diagnostic exceeded or silently shortened its fixed bound: runes=%d", len([]rune(cause)))
	}
}

func assertProjectedDiagnosticFailure(t *testing.T, err error, phase, cause string) Failure {
	t.Helper()
	var failure Failure
	if !errors.As(err, &failure) {
		t.Fatalf("expected typed projected conformance failure, got %v", err)
	}
	if failure.Stage != "verify-result" || failure.Step != "projected-conformance" ||
		failure.Reason != "PROJECTED_CONFORMANCE_FAILED" || failure.UnknownClass != "KNOWN_CONTRADICTION" ||
		failure.NextOperation != "report-counterexample" || len(failure.BlockedBy) != 0 {
		t.Fatalf("native diagnostics changed the admission contract: %#v", failure)
	}
	if len(failure.Diagnostics) != 3 || failure.Diagnostics[1] != "phase="+phase ||
		!strings.HasPrefix(failure.Diagnostics[2], "cause=") || !strings.Contains(failure.Diagnostics[2], cause) {
		t.Fatalf("native failure phase or cause was not retained: %#v", failure.Diagnostics)
	}
	return failure
}
