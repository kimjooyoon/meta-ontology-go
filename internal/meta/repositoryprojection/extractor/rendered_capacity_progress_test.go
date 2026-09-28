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

func TestTypeSafeSuffixCandidateAcceptsStrictIntermediateProgress(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := "package p\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc F() {\n" +
		strings.Repeat("\t_ = fmt.Sprintf(\"%s\", strings.Repeat(\"x\", 1))\n", 100) +
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
		t.Fatal("type-safe suffix fixture lacks F")
	}
	typeEvidence, err := checkTypes(root, "x.go", fset, file, function)
	if err != nil {
		t.Fatal(err)
	}
	existing := functionNames(file)
	var found *suffixCandidate
	for startIndex := range function.Body.List {
		candidate, candidateErr := buildSuffixCandidate([]byte(source), fset, file, function, startIndex, typeEvidence, existing)
		if candidateErr != nil {
			continue
		}
		if candidate != nil && candidate.afterRenderedCapacityOverage > 0 {
			found = candidate
			break
		}
	}
	if found == nil {
		t.Fatal("no type-safe suffix candidate made strict intermediate progress")
	}
	if found.beforeRenderedCapacityOverage <= found.afterRenderedCapacityOverage || found.afterRenderedCapacityOverage <= 0 {
		t.Fatalf("candidate overage before=%d after=%d, want before>after>0", found.beforeRenderedCapacityOverage, found.afterRenderedCapacityOverage)
	}
	if len(found.result) == 0 || len(found.helper) == 0 {
		t.Fatal("accepted suffix candidate did not produce typed source and helper")
	}
	validatedRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(validatedRoot, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validatedRoot, "candidate.go"), found.result, 0o644); err != nil {
		t.Fatal(err)
	}
	validatedSet := token.NewFileSet()
	validatedFile, err := parser.ParseFile(validatedSet, "candidate.go", found.result, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var validatedFunction *ast.FuncDecl
	for _, declaration := range validatedFile.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name != nil && function.Name.Name == "F" {
			validatedFunction = function
			break
		}
	}
	if validatedFunction == nil {
		t.Fatal("validated suffix candidate lacks F")
	}
	if _, err := checkTypes(validatedRoot, "candidate.go", validatedSet, validatedFile, validatedFunction); err != nil {
		t.Fatalf("accepted suffix candidate is not type-safe: %v", err)
	}
}

func TestTypeSafeSuffixIntermediateProgressCompletesPreparation(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := "package p\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc F() {\n" +
		strings.Repeat("\t_ = fmt.Sprintf(\"%s\", strings.Repeat(\"x\", 1))\n", 180) +
		"}\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	prepared, evidence, err := prepareOversizedFunctions(root, "x.go", []byte(source), fset, file)
	if err != nil {
		t.Fatal(err)
	}
	intermediate := false
	for _, item := range evidence {
		if item.Strategy == suffixStrategy && item.AfterRenderedCapacityOverage > 0 && item.BeforeRenderedCapacityOverage > item.AfterRenderedCapacityOverage {
			if item.PreparationProgress == nil || len(item.Obligations) != 0 || len(item.ProofStages) != 0 || len(item.ContractObligations) != 0 {
				t.Fatalf("intermediate suffix evidence=%+v, want typed preparation progress without final proof", item)
			}
			intermediate = true
			break
		}
	}
	if !intermediate {
		t.Fatalf("preparation evidence=%+v, want strict intermediate suffix progress", evidence)
	}
	preparedSet := token.NewFileSet()
	preparedFile, err := parser.ParseFile(preparedSet, "x.go", prepared, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	selection, err := firstOversizedFunction(preparedSet, preparedFile, prepared)
	if err != nil {
		t.Fatal(err)
	}
	if selection.function != nil {
		t.Fatalf("prepared source still has oversized function=%s observations=%+v", selection.function.Name.Name, selection.observations)
	}
}

func TestTypeSafeSuffixProgressFinalizesWithBoundedFinalEvidence(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := "package p\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc F() {\n" +
		strings.Repeat("\t_ = fmt.Sprintf(\"%s\", strings.Repeat(\"x\", 1))\n", 180) +
		"}\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := ExtractWithResult(root, "x.go")
	if err != nil {
		t.Fatal(err)
	}
	intermediate := false
	for _, item := range result.Evidence {
		if item.Strategy != suffixStrategy || item.AfterRenderedCapacityOverage <= 0 {
			continue
		}
		intermediate = true
		if item.PreparationProgress == nil || item.FinalRenderedCapacity == nil || item.FinalRenderedCapacity.Overage != 0 ||
			len(item.Obligations) != 0 || len(item.ProofStages) != 0 || len(item.ContractObligations) != 0 {
			t.Fatalf("progress evidence was promoted into final proof or lost final capacity=%+v", item)
		}
	}
	if !intermediate {
		t.Fatalf("finalized evidence=%+v, want a strict intermediate suffix progress record", result.Evidence)
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
