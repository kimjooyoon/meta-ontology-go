package valueexecution

import (
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
