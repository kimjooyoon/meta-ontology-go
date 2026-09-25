package valueexecution

// ReplayState is the closed-world result of comparing two successful value
// execution receipts. UNKNOWN and REFUTED are deliberately distinct: the
// former means the comparison boundary was not established, while the latter
// is a known contradiction inside an established boundary.
type ReplayState string

const ReplayClosed ReplayState = "CLOSED"
const ReplayUnknown ReplayState = "UNKNOWN"
const ReplayRefuted ReplayState = "REFUTED"

type ReplayComparison struct {
	Schema          string      `json:"schema"`
	State           ReplayState `json:"state"`
	Reason          string      `json:"reason"`
	NextOperation   string      `json:"next_operation"`
	BlockedBy       []string    `json:"blocked_by"`
	SamePlan        bool        `json:"same_plan"`
	SameInput       bool        `json:"same_input"`
	SameExecution   bool        `json:"same_execution"`
	PlanDigest      string      `json:"plan_digest"`
	InputDigest     string      `json:"input_digest"`
	BaselineDigest  string      `json:"baseline_execution_digest"`
	CandidateDigest string      `json:"candidate_execution_digest"`
}

const ReplayComparisonSchema = "gooo/value-execution-replay-comparison/v1"

// CompareReplay compares only receipts produced by the value execution
// boundary. It never declares an improvement; it establishes whether a
// deterministic replay can be reused as evidence for a later decision.
func CompareReplay(baseline, candidate Execution) ReplayComparison {
	comparison := ReplayComparison{
		Schema:          ReplayComparisonSchema,
		SamePlan:        baseline.PlanDigest == candidate.PlanDigest,
		SameInput:       baseline.InputDigest == candidate.InputDigest,
		SameExecution:   baseline.ExecutionDigest == candidate.ExecutionDigest,
		PlanDigest:      baseline.PlanDigest,
		InputDigest:     baseline.InputDigest,
		BaselineDigest:  baseline.ExecutionDigest,
		CandidateDigest: candidate.ExecutionDigest,
	}
	if !replayReceiptShapeValid(baseline) || !replayReceiptShapeValid(candidate) {
		comparison.State = ReplayUnknown
		comparison.Reason = "REPLAY_RECEIPT_INCOMPLETE"
		comparison.NextOperation = "CAPTURE_COMPLETE_EXECUTION_RECEIPT"
		comparison.BlockedBy = []string{"execution_receipt"}
		return comparison
	}
	if executionDigest(baseline) != baseline.ExecutionDigest || executionDigest(candidate) != candidate.ExecutionDigest {
		comparison.State = ReplayRefuted
		comparison.Reason = "REPLAY_RECEIPT_DIGEST_INVALID"
		comparison.NextOperation = "PRESERVE_COUNTEREXAMPLE_AND_OPEN_REPAIR_CANDIDATE"
		return comparison
	}
	if !comparison.SamePlan || !comparison.SameInput {
		comparison.State = ReplayUnknown
		comparison.Reason = "REPLAY_SCOPE_MISMATCH"
		comparison.NextOperation = "CAPTURE_MATCHING_PLAN_AND_INPUT_RECEIPT"
		comparison.BlockedBy = []string{"plan_digest", "input_digest"}
		return comparison
	}
	if !comparison.SameExecution {
		comparison.State = ReplayRefuted
		comparison.Reason = "REPLAY_EXECUTION_DIGEST_MISMATCH"
		comparison.NextOperation = "PRESERVE_COUNTEREXAMPLE_AND_OPEN_REPAIR_CANDIDATE"
		return comparison
	}
	comparison.State = ReplayClosed
	comparison.Reason = "DETERMINISTIC_REPLAY"
	comparison.NextOperation = "REUSE_REPLAY_EVIDENCE"
	return comparison
}

func replayReceiptShapeValid(execution Execution) bool {
	// A failed or in-flight lifecycle receipt is evidence of an attempted
	// execution, never a successful replay boundary. Legacy receipts without a
	// phase remain readable, while new receipts fail closed on explicit failure.
	if execution.Phase == ExecutionPhaseRunning || execution.Phase == ExecutionPhaseFailed {
		return false
	}
	return execution.Scope != "" && validDigest(execution.PlanDigest) &&
		validDigest(execution.InputDigest) && validDigest(execution.ExecutionDigest)
}
