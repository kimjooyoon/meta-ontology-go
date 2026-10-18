package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

const nullaryCallsSource = "package constants\nnamespace constants\n" +
	"entity Integer id \"urn:gooo:type:integer\"\n" +
	"activity Seed() -> Integer id \"urn:nullary:seed\" computes `return 41`\n" +
	"activity Run(Integer) -> Integer id \"urn:nullary:run\" computes `return Seed() + input`\n"

func TestPureNullaryHelperGenerationAndInterpreter(t *testing.T) {
	result, err := Generate("constants.gooo", []byte(nullaryCallsSource), "Run")
	if err != nil {
		t.Fatal(err)
	}
	closure := result.Report.CallClosure
	if closure == nil || len(closure.Activities) != 1 || len(closure.Edges) != 1 ||
		closure.MaxCallsPerInvocation != 1 || closure.Activities[0].ActivityID != "urn:nullary:seed" ||
		!result.Report.RouteEquivalence.Equivalent || !result.Report.DeterministicReplay {
		t.Fatal("nullary helper lost its identity or call cost", result.Report)
	}
	_, passed, err := evaluateIntegerCasesContext(context.Background(), []byte(result.Source), "Run",
		[]IRBodyFillTestCase{{Input: 1, Expected: 42}, {Input: -41, Expected: 0}})
	if err != nil || passed != 2 {
		t.Fatal("nullary helper differs from integer oracle", passed, err)
	}
	changed, err := Generate("constants.gooo", []byte(strings.Replace(nullaryCallsSource, "return 41", "return 42", 1)), "Run")
	if err != nil || changed.Report.CallClosure.Activities[0].ProgramSHA256 == closure.Activities[0].ProgramSHA256 ||
		changed.Report.RouteEquivalence.SourceSemanticDigest == result.Report.RouteEquivalence.SourceSemanticDigest {
		t.Fatal("constant helper change lost its source identity", err)
	}
}

func TestPureNullaryHelperReturnsIndependentRecord(t *testing.T) {
	source := []byte("package records\nnamespace records\n" +
		"entity Result id \"urn:nullary:result\" fields { field value id \"urn:nullary:value\" type integer required one }\n" +
		"activity Empty() -> Result computes `return Result{value: 0}`\n" +
		"activity Run(Result) -> Result computes `let first = Empty(); first.value = input.value; let second = Empty(); return Result{value: first.value + second.value}`\n")
	result, err := Generate("records.gooo", source, "Run")
	if err != nil {
		t.Fatal(err)
	}
	cases := []assemblyspec.ValueCase{{Inputs: `[{"value":7}]`, Expected: `{"value":7}`},
		{Inputs: `[{"value":-2}]`, Expected: `{"value":-2}`}}
	actual, err := evaluateRecordAssembly(context.Background(), []byte(result.Source), "Run", result.Report.RecordTypes, cases)
	if err != nil || len(actual) != 2 || !actual[0].Passed || !actual[1].Passed {
		t.Fatal("record helper captured a caller or shared a mutable result", err, actual)
	}
}

func TestPureNullaryHelperKeepsRejectedCalls(t *testing.T) {
	for _, change := range []struct{ old, next string }{
		{"Seed() + input", "Seed(input)"}, {"Seed() + input", "Missing()"},
		{"return 41", "return input"}, {"return 41", "return Seed()"},
		{"Seed() + input", "len()"}, {"Seed() + input", "int64()"},
	} {
		t.Run(change.next, func(t *testing.T) {
			result, err := Generate("invalid.gooo", []byte(strings.Replace(nullaryCallsSource, change.old, change.next, 1)), "Run")
			if err == nil || result.Source != "" {
				t.Fatal("unsupported nullary call produced output", err)
			}
		})
	}
}

func TestPureNullaryHelperCountsEveryBoundedLoopCall(t *testing.T) {
	for _, calls := range []int{4, 5} {
		body := fmt.Sprintf("let result = input; for i := 0; i < 1024; i++ { result = result + %s }; return result", strings.Repeat("Seed() + ", calls-1)+"Seed()")
		source := strings.Replace(nullaryCallsSource, "return Seed() + input", body, 1)
		result, err := Generate("bounded.gooo", []byte(source), "Run")
		if calls == 4 && (err != nil || result.Report.CallClosure.MaxCallsPerInvocation != 4096) {
			t.Fatal("nullary calls must count toward the existing 4096 bound", err)
		}
		if calls == 5 && (err == nil || result.Source != "" || !strings.Contains(err.Error(), "4096")) {
			t.Fatal("nullary loop escaped the call budget", err)
		}
	}
}

func TestPureNullaryHelperRecordFlowRetainsOriginWithoutInputPort(t *testing.T) {
	source := helperFlowSource(`return "x"`, "Keep()", `"y"`)
	source = []byte(strings.Replace(string(source), "activity Keep(Text)", "activity Keep()", 1))
	flow := exportedFlow(t, source).ValueFlow
	if flow.Status != "RESOLVED" || len(flow.Helpers) != 1 || flow.Helpers[0].ProgramSHA256 == "" {
		t.Fatal("nullary helper lost its source-bound flow", flow)
	}
	helperID, call, literal := flow.Helpers[0].ActivityID, false, false
	for id := range reachableFlowNodes(flow, flow.Choices[0].First) {
		node := flow.Nodes[id-1]
		call = call || node.CalleeID == helperID
		if node.Span.ActivityID == helperID {
			if node.Span.View == "input-port" {
				t.Fatal("nullary helper invented an input dependency", node)
			}
			literal = literal || node.Kind == "literal" && node.Span.View == "computes"
		}
	}
	if !call || !literal {
		t.Fatal("nullary result lost its call or constant origin", call, literal)
	}
}
