package valueexecution

import (
	"context"
	"maps"
)

const ContinuationSchema = "gooo/value-execution-continuation/v1"
const ReasonContinuationInvalid = "VALUE_CONTINUATION_INVALID"
const ReasonContinuationCanceled = "VALUE_CONTINUATION_CANCELED"

// Continuation is detached observation, never a handle for resuming execution.
type Continuation struct {
	Schema              string      `json:"schema"`
	PlanDigest          string      `json:"plan_digest"`
	RuntimePlanDigest   string      `json:"runtime_plan_digest,omitempty"`
	IterationsRequested int         `json:"iterations_requested"`
	IterationsCompleted int         `json:"iterations_completed"`
	FeedbackDeliveries  int         `json:"feedback_deliveries"`
	Executions          []Execution `json:"executions"`
	Failure             *Failure    `json:"failure,omitempty"`
	Digest              string      `json:"digest"`
}

// ExecuteIterations executes a caller-bounded sequence. Only source-declared
// feedback may replace root inputs. Other explicitly supplied roots stay fixed.
// No serialized receipt or caller-created result handle is accepted here.
func (plan Plan) ExecuteIterations(ctx context.Context, rootInputs map[string]int64, iterations int) (trace Continuation, err error) {
	trace = Continuation{Schema: ContinuationSchema, IterationsRequested: iterations}
	defer func() {
		if err != nil {
			failure, ok := FailureOf(err)
			if !ok {
				failure = Failure{Code: ReasonContinuationInvalid, Stage: "CONTINUE", Step: "execute-iterations", Detail: err.Error()}
			}
			trace.Failure = &failure
		}
		trace.Digest = digestValue(trace)
	}()
	if err := plan.validateCompiledAuthority(); err != nil {
		return trace, err
	}
	trace.PlanDigest = planExecutionDigest(plan)
	if iterations < 1 || (iterations > 1 && len(plan.feedbacks) == 0) {
		return trace, failAt(ReasonContinuationInvalid, "PLAN", "require-bounded-feedback", "positive iteration limit and explicit feedback for continuation are required")
	}
	if err := plan.validateFeedbacks(); err != nil {
		return trace, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return plan.executeIterations(ctx, maps.Clone(rootInputs), iterations, trace)
}

func (plan Plan) executeIterations(ctx context.Context, inputs map[string]int64, iterations int, trace Continuation) (Continuation, error) {
	pendingDeliveries := 0
	for index := range iterations {
		if err := ctx.Err(); err != nil {
			return trace, failAt(ReasonContinuationCanceled, "CONTINUE", "check-iteration-context", err.Error())
		}
		execution, values, err := plan.executeIteration(inputs)
		if execution.Scope != "" {
			trace.FeedbackDeliveries += pendingDeliveries
			trace.Executions = append(trace.Executions, execution)
		}
		if err != nil {
			return trace, err
		}
		trace.IterationsCompleted++
		if index+1 < iterations {
			inputs, err = plan.nextIterationInputs(inputs, values)
			if err != nil {
				return trace, err
			}
			pendingDeliveries = len(plan.feedbacks)
		}
	}
	return trace, nil
}

func (plan Plan) validateFeedbacks() error {
	if err := validateExecutionBindings(plan.programs, plan.feedbacks); err != nil {
		return err
	}
	_, incoming, _, err := plan.executionOrder()
	if err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, binding := range plan.feedbacks {
		consumer := binding.Consumer.Activity.Name
		if !binding.Feedback || len(incoming[consumer]) != 0 || seen[consumer] {
			return failAt(ReasonContinuationInvalid, "PLAN", "require-unique-feedback-root", consumer)
		}
		seen[consumer] = true
	}
	return nil
}

func (plan Plan) nextIterationInputs(inputs map[string]int64, values map[string]ProducedResult) (map[string]int64, error) {
	next := maps.Clone(inputs)
	for _, binding := range plan.feedbacks {
		producer, consumer := binding.Producer.Activity.Name, binding.Consumer.Activity.Name
		result, ok := values[producer]
		if !ok {
			return nil, failAt(ReasonBindingResultInvalid, "CONTINUE", "require-produced-feedback-result", producer)
		}
		if err := validateBindingResult(plan.programs[producer], plan.programs[consumer], binding, result); err != nil {
			return nil, failAt(ReasonBindingResultInvalid, "CONTINUE", "validate-feedback-result", err.Error())
		}
		value, err := integerResult(result)
		if err != nil {
			return nil, err
		}
		next[consumer] = value
	}
	return next, nil
}
