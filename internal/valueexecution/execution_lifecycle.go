package valueexecution

// ExecutionPhase is the terminal or in-flight state observed for one value
// plan execution. It is deliberately separate from Plan compilation state so
// an incomplete run cannot be mistaken for a ready plan.
type ExecutionPhase string

const (
	ExecutionPhaseRunning   ExecutionPhase = "RUNNING"
	ExecutionPhaseCompleted ExecutionPhase = "COMPLETED"
	ExecutionPhaseFailed    ExecutionPhase = "FAILED"
)

// ExecutionConditionStatus is an evidence status, not an authorization
// decision. UNKNOWN is preserved when an observation boundary is unavailable.
type ExecutionConditionStatus string

const (
	ExecutionConditionTrue    ExecutionConditionStatus = "TRUE"
	ExecutionConditionFalse   ExecutionConditionStatus = "FALSE"
	ExecutionConditionUnknown ExecutionConditionStatus = "UNKNOWN"
)

const (
	ExecutionConditionPlanReady      = "PlanReady"
	ExecutionConditionExecutionReady = "ExecutionReady"
	ExecutionConditionComplete       = "ExecutionComplete"
)

// ExecutionCondition records one observable boundary of a plan run. Detail is
// diagnostic context and never grants execution or external capability.
type ExecutionCondition struct {
	Type   string                   `json:"type"`
	Status ExecutionConditionStatus `json:"status"`
	Reason string                   `json:"reason"`
	Detail string                   `json:"detail,omitempty"`
}

func executionRunningConditions() []ExecutionCondition {
	return []ExecutionCondition{
		{Type: ExecutionConditionPlanReady, Status: ExecutionConditionTrue, Reason: "COMPILED_PLAN_VALIDATED"},
		{Type: ExecutionConditionExecutionReady, Status: ExecutionConditionFalse, Reason: "EXECUTION_RUNNING"},
		{Type: ExecutionConditionComplete, Status: ExecutionConditionFalse, Reason: "EXECUTION_RUNNING"},
	}
}

func executionLifecycle(err error) (ExecutionPhase, []ExecutionCondition) {
	if err == nil {
		return ExecutionPhaseCompleted, []ExecutionCondition{
			{Type: ExecutionConditionPlanReady, Status: ExecutionConditionTrue, Reason: "COMPILED_PLAN_VALIDATED"},
			{Type: ExecutionConditionExecutionReady, Status: ExecutionConditionTrue, Reason: "EXECUTION_FINISHED"},
			{Type: ExecutionConditionComplete, Status: ExecutionConditionTrue, Reason: "EXECUTION_FINISHED"},
		}
	}
	return ExecutionPhaseFailed, []ExecutionCondition{
		{Type: ExecutionConditionPlanReady, Status: ExecutionConditionTrue, Reason: "COMPILED_PLAN_VALIDATED"},
		{Type: ExecutionConditionExecutionReady, Status: ExecutionConditionFalse, Reason: Reason(err), Detail: err.Error()},
		{Type: ExecutionConditionComplete, Status: ExecutionConditionFalse, Reason: Reason(err), Detail: err.Error()},
	}
}
