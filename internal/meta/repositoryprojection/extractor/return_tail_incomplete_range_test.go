package extractor

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"
)

func TestReturnTailIncompleteRangeTypeEvidenceFailsClosed(t *testing.T) {
	source := "package p\n\nfunc caller() error { return helper() }\n\nfunc helper() error {\n\tvar iterator func(func(int) bool)\n\tfor range iterator {}\n\treturn nil\n}\n"
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var caller, helper *ast.FuncDecl
	var rangeExpression ast.Expr
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil {
			continue
		}
		switch function.Name.Name {
		case "caller":
			caller = function
		case "helper":
			helper = function
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if rangeStatement, ok := node.(*ast.RangeStmt); ok {
					rangeExpression = rangeStatement.X
				}
				return true
			})
		}
	}
	if caller == nil || helper == nil || rangeExpression == nil {
		t.Fatal("incomplete range fixture lacks caller, helper, or range expression")
	}
	evidence, err := checkTypes(root, "x.go", fset, file, caller)
	if err != nil {
		t.Fatal(err)
	}
	delete(evidence.info.Types, rangeExpression)
	helperObject, ok := evidence.info.Defs[helper.Name].(*types.Func)
	if !ok || helperObject == nil {
		t.Fatal("incomplete range fixture lacks typed helper")
	}
	registry := map[string]returnTailHelperProof{
		"helper": returnTailTestHelperProof(t, "helper", evidence, helperObject, helper, "helper-evidence", nil),
	}
	_, err = returnTailCalleeEffects(caller.Body.List, evidence, registry)
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "CALLEE_EFFECTS_UNPROVEN" || failure.UnknownClass != "DIRECT_MISSING" {
		t.Fatalf("incomplete range type evidence error=%v, want fail-closed callee-effects failure", err)
	}
}
