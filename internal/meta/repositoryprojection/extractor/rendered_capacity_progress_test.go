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

func TestRenderedCapacityProgressRequiresStrictOverageDecrease(t *testing.T) {
	cases := []struct {
		name   string
		before int
		after  int
		want   bool
	}{
		{name: "genuine decrease", before: 5, after: 4, want: true},
		{name: "byte-different equal overage", before: 5, after: 5, want: false},
		{name: "byte-different worse overage", before: 5, after: 6, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := renderedCapacitySnapshot{overage: tc.before}
			after := renderedCapacitySnapshot{overage: tc.after}
			if got := renderedCapacityProgress(before, after); got != tc.want {
				t.Fatalf("before=%d after=%d progress=%t, want %t", tc.before, tc.after, got, tc.want)
			}
		})
	}
}

func TestImportHeavyWholeBodyHelperOverCapIsRejected(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := "package p\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc F() {\n" +
		"\t_ = fmt.Sprintf(\"%s\", strings.Repeat(\"x\", 1))\n" +
		strings.Repeat("\t_ = 1\n", 72) +
		"}\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	function, ok := file.Decls[len(file.Decls)-1].(*ast.FuncDecl)
	if !ok || function.Name == nil || function.Name.Name != "F" {
		t.Fatal("whole-body helper fixture lacks F")
	}
	typeEvidence, err := checkTypes(root, "x.go", fset, file, function)
	if err != nil {
		t.Fatal(err)
	}
	_, err = buildSuffixCandidate([]byte(source), fset, file, function, 0, typeEvidence, functionNames(file))
	var contradiction suffixContradiction
	if !errors.As(err, &contradiction) || !strings.Contains(contradiction.Error(), "overage") {
		t.Fatalf("whole-body import-heavy candidate=%v, want strict overage contradiction", err)
	}
}

func TestRenderedCapacityObservesOriginalPaginationSubject(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	logical := "cmd/language-readiness-witness/predecessor-selection/pagination_test.go"
	path := filepath.Join(filepath.Dir(sourceFile), "../../../../", logical)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logical, source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var target *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv == nil && function.Name != nil && function.Name.Name == "TestPaginationFixturesExecuteParserAndHTTPClient" {
			target = function
			break
		}
	}
	if target == nil {
		t.Fatal("original pagination subject was not found")
	}
	observation := observeRenderedCapacity(fset, file, source, target)
	if observation.subject != "func:TestPaginationFixturesExecuteParserAndHTTPClient" || observation.helperStatus == renderedCapacityUnmeasured || observation.helperLines == nil || observation.helperOverage == nil {
		t.Fatalf("pagination observation=%+v, want measured original subject", observation)
	}
	t.Logf("pagination-subject=%s helper_lines=%d helper_overage=%d", observation.subject, *observation.helperLines, *observation.helperOverage)
}
