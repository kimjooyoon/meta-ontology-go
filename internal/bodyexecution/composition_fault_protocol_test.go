package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestArithmeticProtocolRejectsMissingContradictoryAndUnboundOutcomes(t *testing.T) {
	source := []byte(`package protocol
namespace protocol
entity Integer id "protocol://integer"
activity Divide(Integer,Integer) -> Integer computes "return input0 / input1"
activity End(Integer) -> Integer computes "return input + 1"
bind Divide.result -> End.input
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Divide.input0":7,"Divide.input1":0},"expected":{"End":8}}]}`)
	ctx := context.Background()
	prior, err := GenerateComposition(ctx, "protocol.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	graph, err := replayComposition(ctx, "protocol.gooo", source, prior)
	if err != nil {
		t.Fatal(err)
	}
	_, _, sites, err := observedCompositionArtifacts(prior)
	if err != nil || len(sites) != 1 {
		t.Fatal(err, sites)
	}
	valid := `{"schema":"gooo/arithmetic-outcome-observation/v1","outputs":[[null,null]],"faults":[[{"site":0,"left":7,"right":0},null]],"blocked":[[null,["` + graph.nodes[0].ID + `"]]],"calls":[]}`
	_, counts, err := graph.nativeArithmeticTraces([]byte(valid), suite, sites)
	if err != nil || counts.Blocked != 1 || counts.Matched != 0 {
		t.Fatal(err, counts)
	}
	for name, invalid := range map[string]string{
		"site":               strings.Replace(valid, `"site":0`, `"site":9`, 1),
		"missing-site":       strings.Replace(valid, `"site":0,`, ``, 1),
		"missing-left":       strings.Replace(valid, `"left":7,`, ``, 1),
		"missing-zero":       strings.Replace(valid, `,"right":0`, ``, 1),
		"nonzero":            strings.Replace(valid, `"right":0`, `"right":1`, 1),
		"wrong-type":         strings.Replace(valid, `"left":7`, `"left":"7"`, 1),
		"overflow":           strings.Replace(valid, `"left":7`, `"left":9223372036854775808`, 1),
		"duplicate":          strings.Replace(valid, `"right":0`, `"right":0,"right":1`, 1),
		"invent-value":       strings.Replace(valid, `[[null,null]]`, `[[0,null]]`, 1),
		"blocked-value":      strings.Replace(valid, `[[null,null]]`, `[[null,8]]`, 1),
		"dependency":         strings.Replace(valid, graph.nodes[0].ID, "unknown://activity", 1),
		"missing-dependency": strings.Replace(valid, `[[null,["`+graph.nodes[0].ID+`"]]]`, `[[null,null]]`, 1),
		"schema":             strings.Replace(valid, "arithmetic-outcome-observation/v1", "arithmetic-outcome-observation/v0", 1),
		"rows":               strings.Replace(valid, `"outputs":[[null,null]]`, `"outputs":[]`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := graph.nativeArithmeticTraces([]byte(invalid), suite, sites); err == nil {
				t.Fatal("invalid native outcome accepted", invalid)
			}
		})
	}
}

func TestArithmeticObserverLeavesUnknownPanicsTerminal(t *testing.T) {
	root := t.TempDir()
	projection := "package main\nfunc Invoke() int64 { panic(\"unknown native failure\") }\n"
	driver := "package main\nimport(\"encoding/json\";\"os\")\nfunc main(){ value,fault:=GoooObserveArithmetic(Invoke); _=value;_=fault }\n" + calledInputObserver + arithmeticObserver
	var run CompositionRuntime
	executable, err := buildCompositionExecutable(context.Background(), root, projection, driver, nativeTool(), &run)
	if err != nil {
		t.Fatal(err)
	}
	output, observed, err := runCompositionProcess(context.Background(), root, executable, nil)
	if err == nil || len(output) != 0 || observed.ExitCode == nil || *observed.ExitCode == 0 {
		t.Fatal("unknown panic became an observed language fault", err, observed)
	}
}

func TestArithmeticObservationNamesCannotShadowSourceBindings(t *testing.T) {
	source := []byte(`package names
namespace names
entity Integer id "names://integer"
entity GoooArithmeticFault id "names://record" fields { field value id "names://value" type integer required one }
activity Main(Integer,Integer) -> GoooArithmeticFault computes "let GoooObservedArithmetic = input0; let GoooObserveQuotient = GoooObservedArithmetic; return GoooArithmeticFault{value:GoooObserveQuotient / input1}"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main.input0":9007199254740993,"Main.input1":1},"expected":{"Main":{"value":9007199254740993}}}]}`)
	run, prior := executeFaultFixture(t, source, suite)
	if run.FinitePassed != 1 || hasCompositionFault(run) {
		t.Fatal("observer names changed a valid source binding", run)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	<-ctx.Done()
	stopped, err := ExecuteComposition(ctx, "arithmetic.gooo", source, prior, suite, nativeTool())
	if err == nil || stopped.Outcomes != nil || stopped.RuntimeReplayed || len(stopped.Traces) != 0 {
		t.Fatal("cancellation became a measured arithmetic fault", err, stopped)
	}
}

func TestArithmeticRuntimeComparisonBindsFaultsAndOriginalExpectations(t *testing.T) {
	source := []byte("package compare\nnamespace compare\nentity Integer id \"compare://integer\"\nactivity Main(Integer,Integer) -> Integer computes \"return input0 / input1\"\n")
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main.input0":9007199254740993,"Main.input1":0},"expected":{"Main":1}}]}`)
	run, _ := executeFaultFixture(t, source, suite)
	raw, _ := json.Marshal(run)
	for name, change := range map[string]func(*CompositionRuntime){
		"schema":     func(r *CompositionRuntime) { r.Schema = "gooo/body-composition-runtime/v2" },
		"expression": func(r *CompositionRuntime) { r.Traces[0].Deliveries[0].Fault.Site.ExpressionID += "changed" },
		"activity":   func(r *CompositionRuntime) { r.Traces[0].Deliveries[0].Fault.Site.ActivityID = "other" },
		"operator":   func(r *CompositionRuntime) { r.Traces[0].Deliveries[0].Fault.Site.Operator = "%" },
		"left":       func(r *CompositionRuntime) { r.Traces[0].Deliveries[0].Fault.Left-- },
		"right":      func(r *CompositionRuntime) { r.Traces[0].Deliveries[0].Fault.Right = 1 },
		"expected":   func(r *CompositionRuntime) { r.Traces[0].Deliveries[0].Expected = json.RawMessage(`0`) },
		"input":      func(r *CompositionRuntime) { r.Traces[0].Deliveries[0].Inputs[0].Value = json.RawMessage(`0`) },
		"case":       func(r *CompositionRuntime) { r.Traces[0].CaseIndex = 1 },
		"site":       func(r *CompositionRuntime) { r.FaultSites[0].ProjectionSHA256 = "changed" },
		"counts":     func(r *CompositionRuntime) { r.Outcomes.Faulted = 0 },
		"no-counts":  func(r *CompositionRuntime) { r.Outcomes = nil },
		"driver":     func(r *CompositionRuntime) { r.ObservedDriverSHA256 = "changed" },
		"missing":    func(r *CompositionRuntime) { r.Traces[0].Deliveries[0].Fault = nil },
		"runs":       func(r *CompositionRuntime) { r.Runs = r.Runs[:1] },
		"timeout":    func(r *CompositionRuntime) { r.Runs[1].TimedOut = true },
		"stdout":     func(r *CompositionRuntime) { r.Runs[1].StdoutSHA256 = digest([]byte("changed")) },
	} {
		t.Run(name, func(t *testing.T) {
			var saved CompositionRuntime
			if err := json.Unmarshal(raw, &saved); err != nil {
				t.Fatal(err)
			}
			change(&saved)
			if sameJointRuntime(run, saved) {
				t.Fatal("altered arithmetic history accepted")
			}
		})
	}
}
