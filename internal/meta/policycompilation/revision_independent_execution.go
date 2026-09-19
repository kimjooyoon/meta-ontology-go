package policycompilation

// PolicyRevisionIndependentExecution is a caller-owned envelope for a
// separately executed generated judge. It is evidence of execution only; it
// does not grant adoption, mutation, or promotion authority.
type PolicyRevisionIndependentExecution struct {
	Schema               string           `json:"schema"`
	ExecutionMode        string           `json:"execution_mode"`
	SourceDigest         string           `json:"source_digest"`
	SemanticDigest       string           `json:"semantic_digest"`
	GeneratedJudgeDigest string           `json:"generated_judge_digest"`
	InputDigest          string           `json:"input_digest"`
	ResultsDigest        string           `json:"results_digest"`
	RequestedCases       int              `json:"requested_cases"`
	ObservedCases        int              `json:"observed_cases"`
	Results              []DecisionResult `json:"results"`
	ProcessStarted       bool             `json:"process_started"`
	ExitCode             int              `json:"exit_code"`
	ExecutionError       string           `json:"execution_error,omitempty"`
	WallMilliseconds     int64            `json:"wall_ms"`
}
