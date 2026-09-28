package valueexecution

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

const runtimeBindingFanoutFixture = `package runtimebinding
namespace runtimebinding

entity Integer id "gooo://runtime-binding/entity/integer"

activity Produce(Integer) -> Integer computes "int.add:1"
activity ConsumeA(Integer) -> Integer computes "int.add:1"
activity ConsumeB(Integer) -> Integer computes "int.add:1"

bind Produce.result -> ConsumeA.input
bind Produce.result -> ConsumeB.input
`

const typedBooleanCompositionFixture = `package runtimebinding
namespace runtimebinding

entity Integer id "gooo://runtime-binding/entity/integer"
entity Boolean id "gooo://runtime-binding/entity/boolean"

activity IsZero(Integer) -> Boolean computes "int.iszero:0"
activity Not(Boolean) -> Boolean computes "bool.not:0"

bind IsZero.result -> Not.input
`

func TestCompilePlanExecutesActualFanoutAndFreshRuns(t *testing.T) {
	plan, err := CompilePlan("fanout.gooo", []byte(runtimeBindingFanoutFixture))
	if err != nil {
		t.Fatal(err)
	}

	first, err := plan.Execute(map[string]int64{"Produce": 41})
	if err != nil {
		t.Fatal(err)
	}
	if first.ApplyCalls != 3 || first.Deliveries != 2 {
		t.Fatalf("execution counts = applies:%d deliveries:%d, want 3 and 2", first.ApplyCalls, first.Deliveries)
	}
	if first.PlanDigest == "" || first.InputDigest == "" || first.ExecutionDigest == "" {
		t.Fatalf("execution provenance digests are incomplete: %#v", first)
	}
	for activity, want := range map[string]int64{"Produce": 42, "ConsumeA": 43, "ConsumeB": 43} {
		result, ok := first.Results[activity]
		if !ok || result.Value != want || result.ProducerActivity != activity || !validDigest(result.ResultDigest) {
			t.Fatalf("%s result = %#v, want actual value %d", activity, result, want)
		}
	}

	second, err := plan.Execute(map[string]int64{"Produce": 5})
	if err != nil {
		t.Fatal(err)
	}
	if second.Results["ConsumeA"].Value != 7 || second.Results["ConsumeB"].Value != 7 {
		t.Fatalf("fresh plan run mixed stale values: %#v", second.Results)
	}
	if second.PlanDigest != first.PlanDigest || second.InputDigest == first.InputDigest || second.ExecutionDigest == first.ExecutionDigest {
		t.Fatalf("execution provenance did not distinguish plan and input identity: first=%#v second=%#v", first, second)
	}
	third, err := plan.Execute(map[string]int64{"Produce": 41})
	if err != nil || third.ExecutionDigest != first.ExecutionDigest {
		t.Fatalf("identical execution replay changed digest: first=%q third=%q err=%v", first.ExecutionDigest, third.ExecutionDigest, err)
	}
}

func TestCompilePlanPreservesTypedBooleanComposition(t *testing.T) {
	plan, err := CompilePlan("boolean-composition.gooo", []byte(typedBooleanCompositionFixture))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		input       int64
		isZeroValue int64
		notValue    int64
	}{{0, 1, 0}, {7, 0, 1}} {
		execution, err := plan.Execute(map[string]int64{"IsZero": test.input})
		if err != nil {
			t.Fatalf("execute composition(%d): %v", test.input, err)
		}
		isZero, isZeroOK := execution.Results["IsZero"]
		not, notOK := execution.Results["Not"]
		if execution.ApplyCalls != 2 || execution.Deliveries != 1 || !isZeroOK || !notOK ||
			isZero.OutputEntity != BooleanEntity || not.OutputEntity != BooleanEntity ||
			isZero.Value != test.isZeroValue || not.Value != test.notValue ||
			!validDigest(isZero.ResultDigest) || !validDigest(not.ResultDigest) {
			t.Fatalf("typed composition(%d) = %#v, want iszero=%d not=%d", test.input, execution, test.isZeroValue, test.notValue)
		}
	}
}

func TestCompilePlanPreservesBooleanAndComposition(t *testing.T) {
	fixture := `package runtimebinding
namespace runtimebinding

entity Integer id "gooo://runtime-binding/entity/integer"
entity Boolean id "gooo://runtime-binding/entity/boolean"

activity IsZero(Integer) -> Boolean computes "int.iszero:0"
activity Not(Boolean) -> Boolean computes "bool.not:0"
activity And(Boolean) -> Boolean computes "bool.and:1"

bind IsZero.result -> Not.input
bind Not.result -> And.input
`
	plan, err := CompilePlan("boolean-and-composition.gooo", []byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		input       int64
		isZeroValue int64
		notValue    int64
		andValue    int64
	}{{0, 1, 0, 0}, {7, 0, 1, 1}} {
		execution, err := plan.Execute(map[string]int64{"IsZero": test.input})
		if err != nil {
			t.Fatalf("execute Boolean and composition(%d): %v", test.input, err)
		}
		isZero, isZeroOK := execution.Results["IsZero"]
		not, notOK := execution.Results["Not"]
		and, andOK := execution.Results["And"]
		if execution.ApplyCalls != 3 || execution.Deliveries != 2 ||
			len(execution.Activities) != 3 || execution.Activities[0] != "IsZero" ||
			execution.Activities[1] != "Not" || execution.Activities[2] != "And" ||
			!isZeroOK || !notOK || !andOK ||
			isZero.OutputEntity != BooleanEntity || not.OutputEntity != BooleanEntity || and.OutputEntity != BooleanEntity ||
			isZero.Value != test.isZeroValue || not.Value != test.notValue || and.Value != test.andValue ||
			!validDigest(execution.PlanDigest) || !validDigest(execution.InputDigest) ||
			!validDigest(execution.ExecutionDigest) || !validDigest(isZero.ResultDigest) ||
			!validDigest(not.ResultDigest) || !validDigest(and.ResultDigest) {
			t.Fatalf("Boolean and composition(%d) = %#v, want %d -> %d -> %d",
				test.input, execution, test.isZeroValue, test.notValue, test.andValue)
		}
	}
}

func TestCompilePlanPreservesMultiHopTypedBooleanComposition(t *testing.T) {
	fixture := `package runtimebinding
namespace runtimebinding

entity Integer id "gooo://runtime-binding/entity/integer"
entity Boolean id "gooo://runtime-binding/entity/boolean"

activity IsSeven(Integer) -> Boolean computes "int.eq:7"
activity Not(Boolean) -> Boolean computes "bool.not:0"
activity Restore(Boolean) -> Boolean computes "bool.not:0"

bind IsSeven.result -> Not.input
bind Not.result -> Restore.input
`
	plan, err := CompilePlan("boolean-composition-multihop.gooo", []byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		input int64
		want  int64
	}{{7, 1}, {6, 0}} {
		execution, err := plan.Execute(map[string]int64{"IsSeven": test.input})
		if err != nil {
			t.Fatalf("execute multihop composition(%d): %v", test.input, err)
		}
		isSeven, isSevenOK := execution.Results["IsSeven"]
		not, notOK := execution.Results["Not"]
		restored, restoredOK := execution.Results["Restore"]
		if execution.ApplyCalls != 3 || execution.Deliveries != 2 || !isSevenOK || !notOK || !restoredOK ||
			isSeven.OutputEntity != BooleanEntity || not.OutputEntity != BooleanEntity || restored.OutputEntity != BooleanEntity ||
			isSeven.Value != test.want || not.Value != 1-test.want || restored.Value != test.want ||
			!validDigest(isSeven.ResultDigest) || !validDigest(not.ResultDigest) || !validDigest(restored.ResultDigest) {
			t.Fatalf("multi-hop typed composition(%d) = %#v, want %d -> %d -> %d", test.input, execution, test.want, 1-test.want, test.want)
		}
	}
}

func TestPlanFailsClosedWithoutRootInputOrWithUnexpectedInput(t *testing.T) {
	plan, err := CompilePlan("fanout.gooo", []byte(runtimeBindingFanoutFixture))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Execute(nil); Reason(err) != ReasonExternalInputMissing {
		t.Fatalf("missing root input reason = %s, want %s", Reason(err), ReasonExternalInputMissing)
	}
	if _, err := plan.Execute(map[string]int64{"ConsumeA": 41}); Reason(err) != ReasonExternalInputUnexpected {
		t.Fatalf("bound external input reason = %s, want %s", Reason(err), ReasonExternalInputUnexpected)
	}
	if _, err := plan.Execute(map[string]int64{"Unknown": 41}); Reason(err) != ReasonExternalInputUnexpected {
		t.Fatalf("unknown external input reason = %s, want %s", Reason(err), ReasonExternalInputUnexpected)
	}
}

func TestPlanRejectsUncompiledOrTamperedPublicIdentityBeforeApply(t *testing.T) {
	if _, err := (Plan{}).Execute(nil); Reason(err) != ReasonPlanInvalid {
		t.Fatalf("zero plan reason = %s, want %s", Reason(err), ReasonPlanInvalid)
	}

	plan, err := CompilePlan("fanout.gooo", []byte(runtimeBindingFanoutFixture))
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*Plan){
		"source digest":        func(plan *Plan) { plan.SourceDigest = "sha256:" + strings.Repeat("0", 64) },
		"semantic fingerprint": func(plan *Plan) { plan.SemanticFingerprint = "tampered" },
	} {
		t.Run(name, func(t *testing.T) {
			tampered := plan
			mutate(&tampered)
			if _, err := tampered.Execute(map[string]int64{"Produce": 41}); Reason(err) != ReasonPlanInvalid {
				t.Fatalf("tampered plan reason = %s, want %s", Reason(err), ReasonPlanInvalid)
			}
		})
	}
	encoded, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	var publicOnly Plan
	if err := json.Unmarshal(encoded, &publicOnly); err != nil {
		t.Fatal(err)
	}
	if execution, err := publicOnly.Execute(map[string]int64{"Produce": 41}); Reason(err) != ReasonPlanInvalid || execution.ApplyCalls != 0 {
		t.Fatalf("public-only plan execution = %#v / %v, want fail closed before apply", execution, err)
	}
}

func TestPlanPreservesPartialExecutionOnApplyFailure(t *testing.T) {
	source := `package runtimebinding
namespace runtimebinding
entity Integer id "gooo://runtime-binding/entity/integer"
activity Produce(Integer) -> Integer computes "int.add:1"
`
	plan, err := CompilePlan("overflow.gooo", []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	execution, err := plan.Execute(map[string]int64{"Produce": math.MaxInt64})
	if Reason(err) != ReasonIntegerOverflow {
		t.Fatalf("overflow reason = %s, want %s", Reason(err), ReasonIntegerOverflow)
	}
	if execution.ApplyCalls != 1 || len(execution.Activities) != 1 || execution.Activities[0] != "Produce" || len(execution.Results) != 0 {
		t.Fatalf("partial execution = %#v, want one attempted apply and no result", execution)
	}
}

func TestPlanPreservesDeliveryAndProducerResultBeforeConsumerFailure(t *testing.T) {
	plan, err := CompilePlan("fanout.gooo", []byte(runtimeBindingFanoutFixture))
	if err != nil {
		t.Fatal(err)
	}
	execution, err := plan.Execute(map[string]int64{"Produce": math.MaxInt64 - 1})
	if Reason(err) != ReasonIntegerOverflow {
		t.Fatalf("consumer overflow reason = %s, want %s", Reason(err), ReasonIntegerOverflow)
	}
	if execution.ApplyCalls != 2 || execution.Deliveries != 1 || len(execution.Activities) != 2 || execution.Activities[0] != "Produce" || execution.Activities[1] != "ConsumeA" || execution.Results["Produce"].Value != math.MaxInt64 || len(execution.Results) != 1 {
		t.Fatalf("consumer overflow partial execution = %#v", execution)
	}
	if _, ran := execution.Results["ConsumeA"]; ran {
		t.Fatal("failed consumer unexpectedly produced a result")
	}
	if _, ran := execution.Results["ConsumeB"]; ran {
		t.Fatal("second consumer ran after first consumer failure")
	}
}

func TestPlanReportsStableRootInputFailures(t *testing.T) {
	source := `package runtimebinding
namespace runtimebinding
entity Integer id "gooo://runtime-binding/entity/integer"
activity Beta(Integer) -> Integer computes "int.add:1"
activity Alpha(Integer) -> Integer computes "int.add:1"
`
	plan, err := CompilePlan("roots.gooo", []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	first, firstErr := plan.Execute(nil)
	second, secondErr := plan.Execute(nil)
	if Reason(firstErr) != ReasonExternalInputMissing || firstErr.Error() != secondErr.Error() {
		t.Fatalf("unstable missing-root failures = %q / %q", firstErr, secondErr)
	}
	if len(first.Activities) != 0 || first.ApplyCalls != 0 || len(first.Results) != 0 || len(second.Activities) != 0 {
		t.Fatalf("preflight failure executed work: first=%#v second=%#v", first, second)
	}
	unknown, err := plan.Execute(map[string]int64{"Zed": 1, "Aardvark": 1})
	if Reason(err) != ReasonExternalInputUnexpected || !strings.Contains(err.Error(), "Aardvark") || len(unknown.Activities) != 0 {
		t.Fatalf("unstable unknown-root failure = %#v / %v", unknown, err)
	}
}

func TestPlanRejectsTamperedBindingAuthorityBeforeApply(t *testing.T) {
	plan, err := CompilePlan("fanout.gooo", []byte(runtimeBindingFanoutFixture))
	if err != nil {
		t.Fatal(err)
	}

	producer := plan.programs["Produce"]
	calls := instrumentApply(&producer)
	producer.Operation.Activity = "Other"
	plan.programs["Produce"] = producer
	if _, err := plan.Execute(map[string]int64{"Produce": 41}); Reason(err) != ReasonBindingResultInvalid {
		t.Fatalf("tampered authority reason = %s, want %s", Reason(err), ReasonBindingResultInvalid)
	}
	if *calls != 0 {
		t.Fatalf("tampered plan applied producer %d times before rejection", *calls)
	}
}

func TestCompilePlanRejectsUnregisteredOperationAndBindingCycle(t *testing.T) {
	unknown := []byte(`package runtimebinding
namespace runtimebinding
entity Integer id "gooo://runtime-binding/entity/integer"
activity Produce(Integer) -> Integer computes "int.magic:1"
`)
	if _, err := CompilePlan("unknown.gooo", unknown); Reason(err) != ReasonProgramUnknown {
		t.Fatalf("unknown operation reason = %s, want %s", Reason(err), ReasonProgramUnknown)
	}
	cycle := []byte(`package runtimebinding
namespace runtimebinding
entity Integer id "gooo://runtime-binding/entity/integer"
activity Produce(Integer) -> Integer computes "int.add:1"
activity Consume(Integer) -> Integer computes "int.add:1"
bind Produce.result -> Consume.input
bind Consume.result -> Produce.input
`)
	if _, err := CompilePlan("cycle.gooo", cycle); err == nil || !strings.Contains(err.Error(), "runtime binding cycle") {
		t.Fatalf("cycle unexpectedly compiled: %v", err)
	}
}
