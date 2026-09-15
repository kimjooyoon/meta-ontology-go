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
	if !validReplayExecution(baseline) || !validReplayExecution(candidate) {
		comparison.State = ReplayUnknown
		comparison.Reason = "REPLAY_RECEIPT_INCOMPLETE"
		return comparison
	}
	if !comparison.SamePlan || !comparison.SameInput {
		comparison.State = ReplayUnknown
		comparison.Reason = "REPLAY_SCOPE_MISMATCH"
		return comparison
	}
	if !comparison.SameExecution {
		comparison.State = ReplayRefuted
		comparison.Reason = "REPLAY_EXECUTION_DIGEST_MISMATCH"
		return comparison
	}
	comparison.State = ReplayClosed
	comparison.Reason = "DETERMINISTIC_REPLAY"
	return comparison
}

func validReplayExecution(execution Execution) bool {
	return execution.Scope != "" && validDigest(execution.PlanDigest) &&
		validDigest(execution.InputDigest) && validDigest(execution.ExecutionDigest)
}
