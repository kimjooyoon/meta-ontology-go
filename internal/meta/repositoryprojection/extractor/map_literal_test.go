package extractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
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

func TestMapReceiverFusionClosesRemainingCapacity(t *testing.T) {
	root := t.TempDir()
	if err := runtimeWitnessWriteModule(root, mapReceiverWitnessSource(), map[string]string{"support.go": mapReceiverWitnessSupport()}); err != nil {
		t.Fatal(err)
	}
	result, err := ExtractWithResult(root, "x.go")
	if err != nil {
		t.Fatalf("extract receiver witness: %#v", err)
	}
	matched := 0
	for _, item := range result.Evidence {
		if item.Strategy != mapLiteralStrategy {
			continue
		}
		matched++
		if len(item.ProofStages) != 6 || len(item.ContractObligations) != 6 ||
			item.FinalRenderedCapacity == nil || item.FinalRenderedCapacity.Overage != 0 {
			t.Fatalf("caller preparation lost its Gooo obligations or final closure: %+v", item)
		}
		detail := item.ProofStages[1].Detail
		for _, expected := range []string{
			"caller-local-adjacent-pointer-receiver", `"binding":"command"`,
			`"binding_uses":1`, `"before_rendered_overage":1`, `"after_rendered_overage":0`,
			`"source_digest":"sha256:`, `"replacement_digest":"sha256:`,
		} {
			if !strings.Contains(detail, expected) {
				t.Fatalf("caller preparation lost %s: %s", expected, detail)
			}
		}
		prior := item.ProofStages[3].Detail
		if !strings.Contains(prior, "RETURN_TAIL_CONTROL_FLOW_UNSUPPORTED") ||
			!strings.Contains(prior, "function contains unsupported control flow") ||
			strings.Contains(prior, "CALLEE_EFFECTS_UNPROVEN") {
			t.Fatalf("caller preparation misreported the actual prior rejection: %s", prior)
		}
	}
	if matched != 1 {
		t.Fatalf("prepared map witnesses=%d, want 1", matched)
	}
	assertMapLiteralCallerExpressions(t, result)
}

func TestMapLiteralAlternativeRequiresTypedStrategyRejection(t *testing.T) {
	control := "function contains unsupported control flow"
	for _, item := range []struct {
		name   string
		err    error
		reason string
	}{
		{"opaque callee", fail("derive-recipe", "prove-callee-effects", "CALLEE_EFFECTS_UNPROVEN", "DIRECT_MISSING", "restore-callee-evidence", nil), "CALLEE_EFFECTS_UNPROVEN"},
		{"typed control flow", returnTailContradiction(obligationControlFlow, control), "RETURN_TAIL_CONTROL_FLOW_UNSUPPORTED"},
		{"untyped lookalike", knownSuffixContradiction("obligation=" + obligationControlFlow + ": " + control), ""},
		{"free binding", returnTailContradiction(obligationFreeBindings, "binding cannot move"), ""},
		{"missing contract", fail("derive-recipe", "admit-return-tail", "OPERATION_INPUT_CONTRACT_MISSING", "DIRECT_MISSING", "restore-operation-input-contract", nil), ""},
		{"missing type", fail("derive-recipe", "type-check-return-tail", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil), ""},
		{"unknown reason", fail("derive-recipe", "admit-return-tail", "UNKNOWN_FUTURE_REASON", "DIRECT_MISSING", "restore-evidence", nil), ""},
		{"no rejection", nil, ""},
	} {
		t.Run(item.name, func(t *testing.T) {
			previous, eligible := mapLiteralPriorFailure(item.err)
			if eligible != (item.reason != "") {
				t.Fatalf("alternative eligible=%t, expected reason=%q", eligible, item.reason)
			}
			if !eligible {
				return
			}
			if previous.Reason != item.reason {
				t.Fatalf("prior reason=%q, want %q", previous.Reason, item.reason)
			}
			if item.reason == "RETURN_TAIL_CONTROL_FLOW_UNSUPPORTED" {
				detail := strings.Join(previous.Diagnostics, "\n")
				if previous.Stage != "derive-recipe" || previous.Step != "admit-return-tail" ||
					previous.UnknownClass != "KNOWN_CONTRADICTION" ||
					previous.NextOperation != "select-caller-preserving-alternative" ||
					previous.BlockedBy == nil || len(previous.BlockedBy) != 0 ||
					!strings.Contains(detail, "obligation="+obligationControlFlow) ||
					!strings.Contains(detail, "original_rejection="+item.err.Error()) {
					t.Fatalf("prior rejection provenance lost: %#v", previous)
				}
			}
		})
	}
}

func TestMapReceiverFusionRejectsUnsafePlacement(t *testing.T) {
	for _, item := range []struct {
		name, body string
		eligible   bool
	}{
		{"adjacent", "command := factory(); _, _ = command.Run()", true},
		{"extra use", "command := factory(); _, _ = command.Run(); _ = command", false},
		{"intervening effect", "command := factory(); mark(); _, _ = command.Run()", false},
		{"comment", "command := factory()\n// preserve binding\n_, _ = command.Run()", false},
		{"effectful lhs", "sink := map[int]int{}; command := factory(); sink[mark()] = command.Value()", false},
		{"compound assignment", "total := 1; command := factory(); total += command.Value(); _ = total", false},
		{"value receiver initializer", "command := valueFactory(); _, _ = command.Run()", false},
		{"multiple bindings", "command, other := pairFactory(); _, _ = command.Run(); _ = other", false},
		{"deferred use", "command := factory(); _, _ = command.Run(); defer command.Run()", false},
	} {
		t.Run(item.name, func(t *testing.T) {
			source, fset, file, function, evidence := mapReceiverBindingFixture(t, item.body)
			fusion := findMapReceiverFusion(source, fset, file, function, evidence)
			if (fusion != nil) != item.eligible {
				t.Fatalf("eligible=%t, want %t", fusion != nil, item.eligible)
			}
		})
	}
}

func mapReceiverBindingFixture(t *testing.T, body string) ([]byte, *token.FileSet, *ast.File, *ast.FuncDecl, typeEvidence) {
	t.Helper()
	source := []byte("package fixture\n" +
		"type receiver struct{}\n" +
		"func factory() *receiver { return &receiver{} }\n" +
		"func valueFactory() receiver { return receiver{} }\n" +
		"func pairFactory() (*receiver, *receiver) { return factory(), factory() }\n" +
		"func (r *receiver) Run() (string, error) { return \"ok\", nil }\n" +
		"func (r *receiver) Value() int { return 1 }\n" +
		"func mark() int { return 0 }\n" +
		"func Run() {\n" + body + "\n}\n")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	config := types.Config{}
	pkg, err := config.Check("fixture", fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatal(err)
	}
	var function *ast.FuncDecl
	for _, declaration := range file.Decls {
		if candidate, ok := declaration.(*ast.FuncDecl); ok && candidate.Name.Name == "Run" && candidate.Recv == nil {
			function = candidate
		}
	}
	if function == nil {
		t.Fatal("fixture function missing")
	}
	return source, fset, file, function, typeEvidence{pkg: pkg, info: info}
}
