package bodycodegen

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func pureCallEvaluatorFixture(t *testing.T) (integerBodyEvaluator, *ast.FuncDecl) {
	t.Helper()
	source := []byte("package calls\nnamespace calls\nentity Integer id \"gooo://calls/integer\"\nentity Boolean id \"gooo://calls/boolean\"\n" +
		"activity Probe(Integer) -> Boolean computes `return input < 0`\n" +
		"activity Score(Integer) -> Integer computes `if input > 0 || Probe(input) { return 1 }; return 0`\n")
	result, err := Generate("calls.gooo", source, "Score")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "calls.go", result.Source, 0)
	if err != nil {
		t.Fatal(err)
	}
	info := types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}, Types: map[ast.Expr]types.TypeAndValue{}}
	if _, err := new(types.Config).Check("calls", fset, []*ast.File{file}, &info); err != nil {
		t.Fatal(err)
	}
	function, _ := findFunction(file, "Score")
	return integerBodyEvaluator{context: context.Background(), information: info, functions: pureEvaluatorFunctions(file, info)}, function
}

func TestPureActivityCallsShortCircuitAndCancellation(t *testing.T) {
	e, function := pureCallEvaluatorFixture(t)
	input := e.information.Defs[function.Type.Params.List[0].Names[0]]
	for _, value := range []int64{1, -1} {
		e.environment, e.callCount = map[types.Object]any{input: value}, nil
		actual, returned, err := e.evaluateBlock(function.Body)
		if err != nil || !returned || actual != int64(1) {
			t.Fatal(actual, returned, err)
		}
		if (value > 0 && e.callCount != nil) || (value < 0 && (e.callCount == nil || *e.callCount != 1)) {
			t.Fatal("short-circuit call count differs", value, e.callCount)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	e.context = ctx
	if _, _, err := e.evaluateBlock(function.Body); !errors.Is(err, context.Canceled) {
		t.Fatal("pure call evaluation lost cancellation", err)
	}
}
