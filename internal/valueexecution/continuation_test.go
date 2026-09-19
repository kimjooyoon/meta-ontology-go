package valueexecution

import (
	"context"
	"math"
	"reflect"
	"strings"
	"testing"
)

const continuationFixture = `package budget
namespace budget
entity Integer id "budget://entity/integer"
activity ObserveBudget(Integer) -> Integer computes "int.add:0"
activity ConsumeAttempt(Integer) -> Integer computes "int.add:-1"
activity PublishRemaining(Integer) -> Integer computes "int.add:0"
bind ObserveBudget.result -> ConsumeAttempt.input
bind ConsumeAttempt.result -> PublishRemaining.input
feedback PublishRemaining.result -> ObserveBudget.input
`

func continuationPlan(t *testing.T, source string) Plan {
	t.Helper()
	plan, err := CompilePlan("budget.gooo", []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func TestContinuationUsesSourceFeedbackAndPrivateResults(t *testing.T) {
	plan := continuationPlan(t, continuationFixture)
	inputs := map[string]int64{"ObserveBudget": 3}
	trace, err := plan.ExecuteIterations(context.Background(), inputs, 3)
	if err != nil || trace.IterationsCompleted != 3 || trace.FeedbackDeliveries != 2 || len(trace.Executions) != 3 {
		t.Fatalf("trace=%+v err=%v", trace, err)
	}
	for index, execution := range trace.Executions {
		if execution.Results["ObserveBudget"].Value != int64(3-index) || execution.Results["PublishRemaining"].Value != int64(2-index) ||
			execution.ApplyCalls != 3 || execution.Deliveries != 2 {
			t.Fatalf("iteration %d=%+v", index, execution)
		}
	}
	if inputs["ObserveBudget"] != 3 || !validDigest(trace.Digest) {
		t.Fatal("caller input mutated or continuation digest missing")
	}
	replay, err := plan.ExecuteIterations(context.Background(), inputs, 3)
	if err != nil || !reflect.DeepEqual(trace, replay) {
		t.Fatalf("continuation replay changed: %v", err)
	}
}

func TestContinuationFailureStopsBeforeNextFeedback(t *testing.T) {
	plan := continuationPlan(t, continuationFixture)
	trace, err := plan.ExecuteIterations(context.Background(), map[string]int64{"ObserveBudget": math.MinInt64 + 1}, 3)
	if Reason(err) != ReasonIntegerOverflow || trace.IterationsCompleted != 1 || trace.FeedbackDeliveries != 1 || len(trace.Executions) != 2 {
		t.Fatalf("partial trace=%+v err=%v", trace, err)
	}
	if trace.Failure == nil || trace.Failure.Stage != "EXECUTE" || trace.Failure.Step != "apply-int-add" || trace.Executions[1].ApplyCalls != 2 {
		t.Fatalf("missing failure origin: %+v", trace)
	}
	_, handles, err := plan.executeIteration(map[string]int64{"ObserveBudget": math.MinInt64})
	if err == nil || handles != nil {
		t.Fatal("failed iteration released resumable handles")
	}
	recovery, err := plan.ExecuteIterations(context.Background(), map[string]int64{"ObserveBudget": 3}, 2)
	if err != nil || recovery.IterationsCompleted != 2 || recovery.Executions[1].Results["PublishRemaining"].Value != 1 {
		t.Fatalf("failure poisoned independent invocation: %+v %v", recovery, err)
	}
}

func TestContinuationRejectsMissingFeedbackAndInvalidBudgetBeforeExecution(t *testing.T) {
	plain := strings.ReplaceAll(continuationFixture, "feedback PublishRemaining.result -> ObserveBudget.input\n", "")
	for _, tc := range []struct {
		source string
		limit  int
	}{{continuationFixture, 0}, {continuationFixture, -1}, {plain, 2}} {
		plan := continuationPlan(t, tc.source)
		trace, err := plan.ExecuteIterations(context.Background(), map[string]int64{"ObserveBudget": 3}, tc.limit)
		if Reason(err) != ReasonContinuationInvalid || len(trace.Executions) != 0 || trace.FeedbackDeliveries != 0 {
			t.Fatalf("unbounded/undeclared continuation: %+v %v", trace, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	trace, err := continuationPlan(t, continuationFixture).ExecuteIterations(ctx, map[string]int64{"ObserveBudget": 3}, 2)
	if Reason(err) != ReasonContinuationCanceled || len(trace.Executions) != 0 {
		t.Fatalf("canceled continuation executed: %+v %v", trace, err)
	}
}

func TestContinuationRejectsForgedAndOtherSourceHandles(t *testing.T) {
	plan := continuationPlan(t, continuationFixture)
	inputs := map[string]int64{"ObserveBudget": 3}
	if _, err := plan.nextIterationInputs(inputs, map[string]ProducedResult{"PublishRemaining": {}}); err == nil {
		t.Fatal("forged empty result accepted")
	}
	other := continuationPlan(t, strings.ReplaceAll(continuationFixture, "int.add:-1", "int.add:-2"))
	_, results, err := other.executeIteration(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.nextIterationInputs(inputs, results); err == nil {
		t.Fatal("other source's result accepted as this plan's feedback")
	}
}

func TestRecordFeedbackIsExplicitlyUnsupported(t *testing.T) {
	source, _ := recordExample(t)
	source = append(source, []byte("\nfeedback Report.result -> Capture.input\n")...)
	if _, err := CompileRecordPlan("record-feedback.gooo", source); err == nil {
		t.Fatal("record feedback silently interpreted as a same-run binding")
	}
}
