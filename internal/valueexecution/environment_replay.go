package valueexecution

// EnvironmentReplayComparison is the replay boundary that includes the
// measured execution environment. An environment change is UNKNOWN rather
// than REFUTED because it cannot establish a semantic counterexample.
type EnvironmentReplayComparison struct {
	Schema                     string                `json:"schema"`
	State                      ReplayState           `json:"state"`
	Reason                     string                `json:"reason"`
	NextOperation              string                `json:"next_operation"`
	BlockedBy                  []string              `json:"blocked_by"`
	EnvironmentState           EnvironmentTransition `json:"environment_state"`
	SameEnvironment            bool                  `json:"same_environment"`
	BaselineEnvironmentDigest  string                `json:"baseline_environment_digest"`
	CandidateEnvironmentDigest string                `json:"candidate_environment_digest"`
	Replay                     ReplayComparison      `json:"replay"`
}

const EnvironmentReplayComparisonSchema = "gooo/value-execution-environment-replay/v1"

// CompareReplayWithEnvironment establishes replay evidence only when both
// environment observations are complete and unchanged.
func CompareReplayWithEnvironment(
	baseline, candidate Execution,
	baselineEnvironment, candidateEnvironment EnvironmentObservation,
) EnvironmentReplayComparison {
	comparison := EnvironmentReplayComparison{
		Schema:                     EnvironmentReplayComparisonSchema,
		EnvironmentState:           CompareEnvironments(baselineEnvironment, candidateEnvironment),
		BaselineEnvironmentDigest:  baselineEnvironment.Digest,
		CandidateEnvironmentDigest: candidateEnvironment.Digest,
	}
	comparison.SameEnvironment = comparison.EnvironmentState == EnvironmentTransitionUnchanged
	switch comparison.EnvironmentState {
	case EnvironmentTransitionUnknown:
		comparison.State = ReplayUnknown
		comparison.Reason = "REPLAY_ENVIRONMENT_INCOMPLETE"
		comparison.NextOperation = "CAPTURE_COMPLETE_EXECUTION_ENVIRONMENT"
		comparison.BlockedBy = []string{"environment_digest"}
		return comparison
	case EnvironmentTransitionChanged:
		comparison.State = ReplayUnknown
		comparison.Reason = "REPLAY_ENVIRONMENT_CHANGED"
		comparison.NextOperation = "CAPTURE_MATCHING_EXECUTION_ENVIRONMENT"
		comparison.BlockedBy = []string{"environment_digest"}
		return comparison
	}
	comparison.Replay = CompareReplay(baseline, candidate)
	comparison.State = comparison.Replay.State
	comparison.Reason = comparison.Replay.Reason
	comparison.NextOperation = comparison.Replay.NextOperation
	comparison.BlockedBy = comparison.Replay.BlockedBy
	return comparison
}
