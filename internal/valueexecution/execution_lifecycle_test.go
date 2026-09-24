package valueexecution

import "testing"

const lifecycleFixture = `package runtimebinding
namespace runtimebinding

entity Integer id "gooo://runtime-binding/entity/integer"

activity Produce(Integer) -> Integer computes "int.add:1"
activity ConsumeA(Integer) -> Integer computes "int.add:1"
activity ConsumeB(Integer) -> Integer computes "int.add:1"

bind Produce.result -> ConsumeA.input
bind Produce.result -> ConsumeB.input
`

func TestExecutionLifecycleRecordsSuccessfulConditions(t *testing.T) {
	plan, err := CompilePlan("lifecycle.gooo", []byte(lifecycleFixture))
	if err != nil {
		t.Fatal(err)
	}
	execution, err := plan.Execute(map[string]int64{"Produce": 41})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Phase != ExecutionPhaseCompleted {
		t.Fatalf("phase=%q, want %q", execution.Phase, ExecutionPhaseCompleted)
	}
	if condition := lifecycleCondition(execution, ExecutionConditionComplete); condition.Status != ExecutionConditionTrue {
		t.Fatalf("completion condition=%#v, want TRUE", condition)
	}
	if condition := lifecycleCondition(execution, ExecutionConditionExecutionReady); condition.Status != ExecutionConditionTrue {
		t.Fatalf("ready condition=%#v, want TRUE", condition)
	}
	if execution.ExecutionDigest == "" {
		t.Fatal("successful execution did not receive a digest")
	}
}

func TestExecutionLifecyclePreservesFailedExecutionObservation(t *testing.T) {
	plan, err := CompilePlan("lifecycle.gooo", []byte(lifecycleFixture))
	if err != nil {
		t.Fatal(err)
	}
	execution, err := plan.Execute(map[string]int64{})
	if err == nil {
		t.Fatal("missing root input unexpectedly succeeded")
	}
	if execution.Phase != ExecutionPhaseFailed {
		t.Fatalf("phase=%q, want %q", execution.Phase, ExecutionPhaseFailed)
	}
	condition := lifecycleCondition(execution, ExecutionConditionComplete)
	if condition.Status != ExecutionConditionFalse || condition.Reason != ReasonExternalInputMissing {
		t.Fatalf("completion condition=%#v, want failed input condition", condition)
	}
	if execution.ExecutionDigest == "" {
		t.Fatal("failed execution did not preserve a digest")
	}
}

func lifecycleCondition(execution Execution, conditionType string) ExecutionCondition {
	for _, condition := range execution.Conditions {
		if condition.Type == conditionType {
			return condition
		}
	}
	return ExecutionCondition{Type: conditionType, Status: ExecutionConditionUnknown, Reason: "CONDITION_NOT_OBSERVED"}
}
