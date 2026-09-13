package extractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestMapLiteralFallbackKeepsOpaqueCallsInCaller(t *testing.T) {
	source := mapLiteralWitnessSource()
	root := t.TempDir()
	if err := runtimeWitnessWriteModule(root, source, map[string]string{"support.go": mapLiteralWitnessSupport()}); err != nil {
		t.Fatal(err)
	}
	result, err := ExtractWithResult(root, "x.go")
	if err != nil {
		t.Fatal(err)
	}
	var matched int
	for _, item := range result.Evidence {
		if item.Strategy != mapLiteralStrategy {
			continue
		}
		matched++
		if len(item.ProofStages) != 6 || len(item.ContractObligations) != 6 || item.ContractSourceDigest == "" || item.ContractSemanticDigest == "" {
			t.Fatalf("map constructor lost Gooo proof obligations: %+v", item)
		}
		if !strings.Contains(item.ProofStages[3].Detail, "CALLEE_EFFECTS_UNPROVEN") {
			t.Fatal("alternative erased the previous return-tail failure")
		}
		if item.BeforeRenderedCapacityOverage <= item.AfterRenderedCapacityOverage || item.FinalRenderedCapacity == nil || item.FinalRenderedCapacity.Overage != 0 {
			t.Fatalf("rendered capacity did not close: %+v", item)
		}
	}
	if matched != 1 {
		t.Fatalf("map constructor witnesses=%d, want 1", matched)
	}
	assertMapLiteralCallerExpressions(t, result)
}

func assertMapLiteralCallerExpressions(t *testing.T, result Result) {
	t.Helper()
	var callerCalls int
	for path, source := range result.Generated {
		file, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					if strings.Contains(function.Name.Name, "ExtractedMap") {
						t.Fatal("constructor contains a moved call")
					}
					if callee, ok := call.Fun.(*ast.Ident); ok && callee.Name == "next" && function.Name.Name == "Run" {
						callerCalls++
					}
				}
				return true
			})
		}
	}
	if callerCalls != 16 {
		t.Fatalf("caller value evaluations=%d, want 16", callerCalls)
	}
}

func TestMapLiteralConstantKeyBoundary(t *testing.T) {
	for _, item := range []struct {
		name, expression string
		eligible         bool
	}{
		{"constant keys", "map[string]any{\"a\": 1, \"b\": nil}", true},
		{"dynamic key", "map[string]any{key(): 1, \"b\": 2}", false},
		{"duplicate key", "map[string]any{\"a\": 1, \"a\": 2}", false},
		{"escaped duplicate", "map[string]any{\"a\": 1, \"\\x61\": 2}", false},
	} {
		t.Run(item.name, func(t *testing.T) {
			expression, err := parser.ParseExpr(item.expression)
			if err != nil {
				t.Fatal(err)
			}
			_, ok := mapLiteralKeys(expression.(*ast.CompositeLit))
			if ok != item.eligible {
				t.Fatalf("eligible=%t, want %t", ok, item.eligible)
			}
		})
	}
}
