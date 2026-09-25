package linecaps

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCollapseAssignReturnPreservesExistingVariableWrites(t *testing.T) {
	operators := []string{"=", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<=", ">>=", "&^=", ":="}
	if len(operators) != 13 {
		t.Fatal("assignment-form regression denominator changed")
	}
	for _, operator := range operators {
		t.Run(operator, func(t *testing.T) {
			source := fmt.Sprintf("package p\nvar value = 8\nfunc change() int { value %s 1; return value }\n", operator)
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", source, 0)
			if err != nil {
				t.Fatal(err)
			}
			function := file.Decls[1].(*ast.FuncDecl)
			first := function.Body.List[0]
			rule, _, _, _ := classifyRefactorCandidate(function.Body)
			want := operator == ":="
			if (rule == RuleRefactorAssign) != want {
				t.Fatalf("operator=%s rule=%s: only a new local binding may be collapsed", operator, rule)
			}
			if changed := CollapseAssignReturn(fset, function, nil); changed != want {
				t.Fatalf("operator=%s changed=%t want=%t", operator, changed, want)
			}
			if !want && (len(function.Body.List) != 2 || function.Body.List[0] != first) {
				t.Fatalf("operator=%s lost an existing-variable write", operator)
			}
			if want {
				statement, ok := function.Body.List[0].(*ast.ReturnStmt)
				if len(function.Body.List) != 1 || !ok || len(statement.Results) != 1 {
					t.Fatal("new local binding did not become one return")
				}
			}
		})
	}
}
