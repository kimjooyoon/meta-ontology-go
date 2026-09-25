package extractor

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStrictRenderedCapacityProgressRequiresTotalOverageDecrease(t *testing.T) {
	cases := []struct {
		name   string
		before renderedCapacitySnapshot
		after  renderedCapacitySnapshot
		want   bool
	}{
		{name: "strict decrease", before: renderedCapacitySnapshot{overage: 5}, after: renderedCapacitySnapshot{overage: 4}, want: true},
		{name: "equal overage", before: renderedCapacitySnapshot{overage: 5}, after: renderedCapacitySnapshot{overage: 5}, want: false},
		{name: "increased overage", before: renderedCapacitySnapshot{overage: 5}, after: renderedCapacitySnapshot{overage: 6}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := strictRenderedCapacityProgress(tc.before, tc.after); got != tc.want {
				t.Fatalf("strictRenderedCapacityProgress(%+v, %+v)=%v, want %v", tc.before, tc.after, got, tc.want)
			}
		})
	}
}

func TestSuffixCandidateRejectsFullRenderedHelperOverCap(t *testing.T) {
	root := t.TempDir()
	logical := "x.go"
	source := "package p\n\nimport (\n" +
		"\t\"bytes\"\n" +
		"\t\"crypto/sha256\"\n" +
		"\t\"encoding/hex\"\n" +
		"\t\"encoding/json\"\n" +
		"\t\"fmt\"\n" +
		"\t\"net/http\"\n" +
		"\t\"net/url\"\n" +
		"\t\"path/filepath\"\n" +
		"\t\"reflect\"\n" +
		"\t\"runtime\"\n" +
		"\t\"sort\"\n" +
		"\t\"strconv\"\n" +
		"\t\"strings\"\n" +
		"\t\"time\"\n" +
		"\t\"unicode\"\n" +
		")\n\nfunc F() {\n" +
		"\t_ = bytes.Clone(nil)\n" +
		"\t_ = sha256.Size\n" +
		"\t_ = hex.EncodedLen(1)\n" +
		"\t_ = json.Valid(nil)\n" +
		"\t_ = fmt.Sprint(1)\n" +
		"\t_ = http.MethodGet\n" +
		"\t_ = url.URL{}\n" +
		"\t_ = filepath.Separator\n" +
		"\t_ = reflect.TypeOf(nil)\n" +
		"\t_ = runtime.GOOS\n" +
		"\t_ = sort.Ints\n" +
		"\t_ = strconv.IntSize\n" +
		"\t_ = strings.TrimSpace(\"\")\n" +
		"\t_ = time.Second\n" +
		"\t_ = unicode.IsLetter('a')\n" +
		strings.Repeat("\t_ = 1\n", 54) +
		"}\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, logical), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logical, []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	function, ok := file.Decls[len(file.Decls)-1].(*ast.FuncDecl)
	if !ok || function.Name.Name != "F" {
		t.Fatalf("function=%T, want F declaration", file.Decls[len(file.Decls)-1])
	}
	evidence, err := checkTypes(root, logical, fset, file, function)
	if err != nil {
		t.Fatal(err)
	}

	first := function.Body.List[0]
	last := function.Body.List[len(function.Body.List)-1]
	start := fset.Position(first.Pos()).Offset
	end := fset.Position(last.End()).Offset
	helper, err := renderSuffixHelper(fset, "FExtractedSuffix01", nil, []byte(source)[start:end])
	if err != nil {
		t.Fatal(err)
	}
	if physicalLines(helper) > functionLineLimit {
		t.Fatalf("body-only suffix helper=%d lines, want within cap", physicalLines(helper))
	}
	if declarationLines(fset, function) > functionLineLimit {
		t.Fatalf("source function=%d lines, want body within cap", declarationLines(fset, function))
	}
	full, err := renderedDeclarationHelper(fset, file, []byte(source), function)
	if err != nil {
		t.Fatal(err)
	}
	if physicalLines(full) <= functionLineLimit {
		t.Fatalf("full rendered helper=%d lines, want over cap", physicalLines(full))
	}

	_, err = buildSuffixCandidate([]byte(source), fset, file, function, 0, evidence, map[string]bool{})
	var contradiction suffixContradiction
	if !errors.As(err, &contradiction) || !strings.Contains(err.Error(), "overage") {
		t.Fatalf("err=%v, want full-rendered-capacity contradiction", err)
	}
}

func TestPaginationGenericExtractionTerminatesWithUnprovenCallbackIdentity(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "../../../.."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repository root unavailable: %v", err)
	}
	result, err := ExtractWithResult(root, "cmd/language-readiness-witness/predecessor-selection/pagination_test.go")
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "CALLBACK_ENCLOSING_IDENTITY_UNPROVEN" || failure.UnknownClass != "UNBOUNDED" ||
		failure.Stage != "derive-recipe" || failure.Step != "preserve-callback-identity" ||
		failure.NextOperation != "prove-callback-observability" || len(failure.BlockedBy) != 0 {
		t.Fatalf("result=%+v err=%v, want an unproven preservation obligation", result, err)
	}
	if len(result.Generated) != 0 {
		t.Fatalf("generated=%d, want no conversion output without preservation evidence", len(result.Generated))
	}
	for _, prefix := range []string{"measurement=MEASURED", "helper_lines=", "helper_overage=", "suffix_candidate_index=", "suffix_candidates_attempted=", "suffix_candidates_rejected=", "suffix_candidates_unproven="} {
		found := false
		for _, diagnostic := range failure.Diagnostics {
			if strings.HasPrefix(diagnostic, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("diagnostics=%v, want actual pagination observation field %q", failure.Diagnostics, prefix)
		}
	}
}
