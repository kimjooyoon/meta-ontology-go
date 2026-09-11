package policycompilation

const PolicyRevisionObservationSchema = "gooo/meta-policy-revision-observation/v1"

// PolicyRevisionCasePair preserves two caller-owned observation snapshots.
// A changed policy is never permission to rebind either snapshot's digests.
type PolicyRevisionCasePair struct {
	Baseline  Case `json:"baseline"`
	Candidate Case `json:"candidate"`
}

type PolicyRevisionObservationRequest struct {
	ExpectedSourceDigest string                   `json:"expected_source_digest"`
	Condition            string                   `json:"condition"`
	FromDecision         string                   `json:"from_decision"`
	ToDecision           string                   `json:"to_decision"`
	Cases                []PolicyRevisionCasePair `json:"cases"`
}

type PolicyRevisionPending struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

// SourceResults are producer-side interpretation, not independent evidence.
type PolicyRevisionExecution struct {
	GeneratedJudgeSource string           `json:"generated_judge_source"`
	GeneratedJudgeDigest string           `json:"generated_judge_digest"`
	DeclaredInputs       []Case           `json:"declared_inputs"`
	SourceResults        []DecisionResult `json:"source_results"`
	FirstResults         []DecisionResult `json:"first_results"`
	ReplayResults        []DecisionResult `json:"replay_results"`
	Complete             bool             `json:"complete"`
	ExecutionError       string           `json:"execution_error,omitempty"`
	WallMilliseconds     int64            `json:"wall_ms"`
}

type PolicyRevisionObservedTransition struct {
	CaseID                      string `json:"case_id"`
	InputsIdentical             bool   `json:"inputs_identical"`
	BaselineCondition           string `json:"baseline_condition"`
	CandidateCondition          string `json:"candidate_condition"`
	BaselineDecision            string `json:"baseline_decision"`
	CandidateDecision           string `json:"candidate_decision"`
	RequestedTransitionObserved bool   `json:"requested_transition_observed"`
	CausalAttribution           string `json:"causal_attribution"`
}

type PolicyRevisionObservationCounts struct {
	RequestedCasePairs           int `json:"requested_case_pairs"`
	ObservedCasePairs            int `json:"observed_case_pairs"`
	SameInputObservedPairs       int `json:"same_input_observed_pairs"`
	RequestedTransitionsObserved int `json:"requested_transitions_observed"`
	SourceComparisons            int `json:"source_comparisons"`
	SourceMismatches             int `json:"source_mismatches"`
	ReplayComparisons            int `json:"replay_comparisons"`
	ReplayMismatches             int `json:"replay_mismatches"`
	FailedBatches                int `json:"failed_batches"`
}

// PolicyRevisionObservation is bounded execution evidence, not adoption.
// A generated result cannot supply the missing independent validation.
type PolicyRevisionObservation struct {
	Schema                string                             `json:"schema"`
	SourceFile            string                             `json:"source_file"`
	RequestDigest         string                             `json:"canonical_request_digest"`
	RequestArtifactDigest string                             `json:"request_artifact_digest,omitempty"`
	Request               PolicyRevisionObservationRequest   `json:"request"`
	OriginalPolicy        CompiledPolicy                     `json:"original_policy"`
	CandidatePolicy       CompiledPolicy                     `json:"candidate_policy"`
	CandidateSource       string                             `json:"candidate_source"`
	ChangedCoordinates    []string                           `json:"changed_coordinates"`
	Baseline              PolicyRevisionExecution            `json:"baseline"`
	Candidate             PolicyRevisionExecution            `json:"candidate"`
	Transitions           []PolicyRevisionObservedTransition `json:"transitions"`
	Counts                PolicyRevisionObservationCounts    `json:"counts"`
	ExecutionStatus       string                             `json:"execution_status"`
	ExecutionConformance  string                             `json:"execution_conformance"`
	Admission             PolicyRevisionPending              `json:"admission"`
	Pending               []PolicyRevisionPending            `json:"pending"`
	InputProvenance       string                             `json:"input_provenance"`
	RepositoryObservation string                             `json:"repository_observation"`
	Improvement           string                             `json:"improvement"`
	MutationAuthority     int                                `json:"mutation_authority"`
	PromotionAuthority    int                                `json:"promotion_authority"`
}
