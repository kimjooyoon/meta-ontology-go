package valueexecution

import (
	"strings"
	"testing"
)

func TestRuntimePlanDigestBindsExecutionAndContinuationEvidence(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	execution := Execution{
		Scope:      RegisteredValueOperationScope,
		PlanDigest: digest,
		Phase:      ExecutionPhaseCompleted,
	}
	boundExecution, err := execution.BindRuntimePlanDigest(digest)
	if err != nil {
		t.Fatalf("BindRuntimePlanDigest() error = %v", err)
	}
	if boundExecution.RuntimePlanDigest != digest || boundExecution.ExecutionDigest != executionDigest(boundExecution) {
		t.Fatalf("bound execution evidence = %#v", boundExecution)
	}
	trace := Continuation{
		Schema:              ContinuationSchema,
		PlanDigest:          digest,
		IterationsRequested: 1,
		IterationsCompleted: 1,
		Executions:          []Execution{execution},
	}
	boundTrace, err := trace.BindRuntimePlanDigest(digest)
	if err != nil {
		t.Fatalf("Continuation.BindRuntimePlanDigest() error = %v", err)
	}
	expectedTraceDigest := boundTrace
	expectedTraceDigest.Digest = ""
	if boundTrace.RuntimePlanDigest != digest || len(boundTrace.Executions) != 1 ||
		boundTrace.Executions[0].RuntimePlanDigest != digest || boundTrace.Digest != digestValue(expectedTraceDigest) {
		t.Fatalf("bound continuation evidence = %#v", boundTrace)
	}
	if _, err := execution.BindRuntimePlanDigest("not-a-digest"); err == nil {
		t.Fatal("invalid runtime plan digest was accepted")
	}
}
