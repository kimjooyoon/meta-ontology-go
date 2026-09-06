package extractor

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCapacityFailureIncludesMeasuredArtifactOverage(t *testing.T) {
	source := []byte("package p\n\n" + strings.Repeat("// documentation\n", functionLineLimit+1) + "func F() {}\n")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	function := file.Decls[0].(*ast.FuncDecl)
	observation := observeRenderedCapacity(fset, file, source, function)
	if observation.helperLines == nil || *observation.helperLines <= functionLineLimit {
		t.Fatalf("observation=%+v, want a measured over-cap helper", observation)
	}
	original := fail("derive-recipe", "select-safe-suffix", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-counterexample", nil)
	var failure Failure
	if !errors.As(withRenderedCapacityDiagnostics(original, []renderedCapacityObservation{observation}), &failure) {
		t.Fatal("capacity failure lost its structured classification")
	}
	for _, want := range []string{
		"measurement=MEASURED", "subject=func:F", "source_digest=" + proofDigest(source),
		"function_status=WITHIN_CAP", "helper_status=OVER_CAP", "function_overage=0",
		"helper_measurement_scope=" + renderedCapacityHelperMeasurementScope,
		fmt.Sprintf("function_line_limit=%d", functionLineLimit),
		fmt.Sprintf("helper_lines=%d", *observation.helperLines),
		fmt.Sprintf("helper_overage=%d", *observation.helperLines-functionLineLimit),
	} {
		if !capacityDiagnosticPresent(failure.Diagnostics, want) {
			t.Fatalf("diagnostics=%v, missing exact observation %q", failure.Diagnostics, want)
		}
	}
	if failure.Stage != "derive-recipe" || failure.Step != "select-safe-suffix" || failure.UnknownClass != "KNOWN_CONTRADICTION" {
		t.Fatalf("failure=%+v, diagnostics must not change the decision", failure)
	}
}

func TestCapacityFailureDoesNotInventUnmeasuredHelperValues(t *testing.T) {
	source := []byte("package p\n\nfunc F() {}\n")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	renderer := func(_ *token.FileSet, _ *ast.File, _ []byte, _ *ast.FuncDecl) ([]byte, error) {
		return nil, fail("observe-plan", "render-capacity", "PREFLIGHT_RENDER_FAILED", "DIRECT_MISSING", "restore-render-evidence", nil)
	}
	observation := observeRenderedCapacityWithRenderer(fset, file, source, file.Decls[0].(*ast.FuncDecl), renderer)
	var failure Failure
	if !errors.As(renderedCapacityObservationFailure(observation), &failure) {
		t.Fatal("missing structured measurement failure")
	}
	if failure.UnknownClass != "DIRECT_MISSING" || !capacityDiagnosticPresent(failure.Diagnostics, "measurement=UNMEASURED") {
		t.Fatalf("failure=%+v, want unchanged unknown measurement", failure)
	}
	for _, diagnostic := range failure.Diagnostics {
		if strings.HasPrefix(diagnostic, "helper_lines=") || strings.HasPrefix(diagnostic, "helper_overage=") {
			t.Fatalf("unmeasured helper received a numeric observation: %s", diagnostic)
		}
	}
}

func TestCapacityFailureRetainsEveryRejectedSuffixInAttemptOrder(t *testing.T) {
	root := t.TempDir()
	source := "package p\n\nfunc F() {\n" +
		"\tdefer func() {\n" + strings.Repeat("\t\t_ = 1\n", functionLineLimit) + "\t}()\n" +
		"\tdefer func() {\n" + strings.Repeat("\t\t_ = 2\n", functionLineLimit) + "\t}()\n}\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "fixture.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := ExtractWithResult(root, "fixture.go")
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "NO_SAFE_DECLARATION_CAPACITY" || len(result.Generated) != 0 {
		t.Fatalf("result=%+v err=%v, want a non-generating capacity refusal", result, err)
	}
	var candidates []string
	for _, diagnostic := range failure.Diagnostics {
		if strings.HasPrefix(diagnostic, "suffix_candidate_index=") {
			candidates = append(candidates, diagnostic)
		}
	}
	if len(candidates) != 2 {
		t.Fatalf("candidates=%v, want exactly two attempted suffixes", candidates)
	}
	for order, candidate := range candidates {
		if !strings.HasPrefix(candidate, fmt.Sprintf("suffix_candidate_index=%d;", 1-order)) ||
			!strings.Contains(candidate, "statement_start=fixture.go:") ||
			!strings.Contains(candidate, "statement_end=fixture.go:") ||
			!strings.Contains(candidate, `rejection="suffix control-flow or scope invariant is not preserved"`) {
			t.Fatalf("candidate=%q, want source-bound rejection in actual attempt order", candidate)
		}
	}
	for _, want := range []string{"suffix_candidates_attempted=2", "suffix_candidates_rejected=2"} {
		if !capacityDiagnosticPresent(failure.Diagnostics, want) {
			t.Fatalf("diagnostics=%v, missing exact denominator %q", failure.Diagnostics, want)
		}
	}
}

func capacityDiagnosticPresent(diagnostics []string, want string) bool {
	return slices.Contains(diagnostics, want)
}
