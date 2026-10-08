package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestArithmeticOutcomesPreserveBranchesAndExactIntegers(t *testing.T) {
	for _, test := range []struct {
		name, body, result, expected string
	}{
		{"if", "if input1 == 0 { return 0 }; return Divide(input0,input1)", "Integer", "[0,-9223372036854775808,9007199254740993,-2]"},
		{"and", "return input1 != 0 && Divide(input0,input1) > 0", "Boolean", "[false,false,true,false]"},
		{"or", "return input1 == 0 || Divide(input0,input1) > 0", "Boolean", "[true,false,true,false]"},
		{"remainder", "if input1 == 0 { return 0 }; return input0 % input1", "Integer", "[0,0,0,-1]"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := []byte(fmt.Sprintf(`package arithmetic
namespace arithmetic
entity Integer id "arithmetic://integer"
entity Boolean id "arithmetic://boolean"
activity Divide(Integer,Integer) -> Integer computes "return input0 / input1"
activity Main(Integer,Integer) -> %s computes %q
`, test.result, test.body))
			var expected []json.RawMessage
			if err := json.Unmarshal([]byte(test.expected), &expected); err != nil {
				t.Fatal(err)
			}
			suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
			for i, pair := range [][2]string{{"5", "0"}, {"-9223372036854775808", "-1"}, {"9007199254740993", "1"}, {"-7", "3"}} {
				suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{
					"Main.input0": json.RawMessage(pair[0]), "Main.input1": json.RawMessage(pair[1])}, Expected: map[string]json.RawMessage{"Main": expected[i]}})
			}
			run, prior := executeFaultFixture(t, source, suite)
			if run.FinitePassed != 4 || hasCompositionFault(run) || !run.RuntimeReplayed || len(run.FaultSites) != 1 {
				t.Fatal("branch, short circuit or integer behavior changed", run)
			}
			if strings.Contains(prior.Source, "GoooObserve") || strings.Contains(prior.Driver, "GoooArithmetic") {
				t.Fatal("observation rewrote saved source")
			}
			assertFaultJSON(t, run, 0, 0, 4)
		})
	}
}

func executeFaultFixture(t *testing.T, source []byte, suite CompositionCases) (CompositionRuntime, Composition) {
	t.Helper()
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "arithmetic.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(ctx, "arithmetic.gooo", source, prior, suite, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	return run, prior
}

func TestArithmeticOutcomesRetainFirstReachedOperation(t *testing.T) {
	for _, test := range []struct{ body, operator string }{
		{"return input0 / input1", "/"},
		{"return input0 % input1", "%"},
		{"return input0 / (input0 % input1)", "%"},
		{"return (input0 / input1) % input1", "/"},
		{"let ignored = input0 / input1; return input0", "/"},
	} {
		t.Run(test.body, func(t *testing.T) {
			source := []byte("package arithmetic\nnamespace arithmetic\nentity Integer id \"arithmetic://integer\"\nactivity Main(Integer,Integer) -> Integer computes " + fmt.Sprintf("%q", test.body))
			suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main.input0":9007199254740993,"Main.input1":0},"expected":{"Main":1}}]}`)
			run, _ := executeFaultFixture(t, source, suite)
			fault := run.Traces[0].Deliveries[0].Fault
			if fault == nil || fault.Kind != "ZERO_DIVISOR" || fault.Site.Operator != test.operator ||
				fault.Left != 9007199254740993 || fault.Right != 0 || fault.Site.ExpressionID == "" || fault.Site.ProjectionSHA256 != run.GeneratedSHA256 {
				t.Fatal("first reached operation or exact operands lost", fault)
			}
			assertFaultJSON(t, run, 1, 0, 0)
		})
	}
}

func TestJointAllCallerFaultsRemainPartialAndReplay(t *testing.T) {
	source := []byte(`package allfaults
namespace allfaults
entity Integer id "allfaults://integer"
activity Divide(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__ + __GOOO_BODY_HOLE_bias__" assembling {
 source_fill intent "Return one." {
  hole "value"
  hole "bias"
  candidate "first" { fill "value" "input / input" fill "bias" "0" }
  candidate "second" { fill "value" "(input * 2) / input - 1" fill "bias" "0" }
 }
 case "1" -> "1"
}
activity Main(Integer) -> Integer computes "return Divide(input)"
`)
	cases := jointCases(t, `[{"inputs":{"Main":0},"expected":{"Main":1}}]`)
	r, err := ConstructJointComposition(context.Background(), "allfaults.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil || len(r.Attempts) != 2 || r.Decision != "PARTIAL_FINITE" || r.StopReason != "DECLARED_SPACE_EXHAUSTED" {
		t.Fatal("all-fault space did not preserve a finite partial result", err, r)
	}
	for _, a := range r.Attempts {
		assertFaultJSON(t, a.Runtime, 1, 0, 0)
	}
	replayed, err := ReplayJointComposition(context.Background(), "allfaults.gooo", source, r, cases, nativeTool())
	if err != nil || !replayed.ConstructionReplayed || replayed.NewModelCalls != 0 || replayed.Runtime.FinitePassed != 0 {
		t.Fatal("all-fault history could not replay", err, replayed)
	}
}
