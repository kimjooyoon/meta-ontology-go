package main

// Consumer-owned JSON declarations preserve the v1 wire contract without
// linking the producer compiler or evaluator into the observer.
const revisionWireSchema = "gooo/meta-policy-revision-observation/v1"

type revisionWireRule struct {
	ActivityID    string `json:"activity_id"`
	ActivityName  string `json:"activity_name"`
	Role          string `json:"role"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          int    `json:"step"`
	Reason        string `json:"reason"`
	Claim         string `json:"claim"`
}

type revisionWirePolicy struct {
	Schema         string                `json:"schema"`
	PolicyID       string                `json:"policy_id"`
	Package        string                `json:"package"`
	Namespace      string                `json:"namespace"`
	SourceDigest   string                `json:"source_digest"`
	SemanticDigest string                `json:"semantic_digest"`
	Denominator    int                   `json:"fixed_denominator"`
	Rules          []revisionWireRule    `json:"rules"`
	Reduction      revisionWireReduction `json:"decision_reduction"`
	Structure      revisionWireStructure `json:"structure"`
}

type revisionWireStructure struct {
	GrammarNodeKinds        int            `json:"grammar_node_kinds"`
	ASTNodeCounts           map[string]int `json:"ast_node_counts"`
	IRNodeCounts            map[string]int `json:"ir_node_counts"`
	TransitionBindings      int            `json:"transition_bindings"`
	EvidenceBindings        int            `json:"evidence_bindings"`
	ResolutionBindings      int            `json:"resolution_bindings"`
	MarkerOccurrencesBefore int            `json:"marker_occurrences_before"`
	MarkerOccurrencesAfter  int            `json:"marker_occurrences_after"`
	MarkerImprovement       string         `json:"marker_improvement"`
}

type revisionWireReduction struct {
	Schema string                `json:"schema"`
	Rules  []revisionReceiptRule `json:"rules"`
}

type revisionWireCase struct {
	ID                           string `json:"id"`
	ValidatorExpectation         string `json:"validator_expectation"`
	EvidenceClass                string `json:"evidence_class"`
	Provenance                   string `json:"provenance"`
	ProducerAvailable            bool   `json:"producer_available"`
	ConsumerAvailable            bool   `json:"consumer_available"`
	ObservedSourceDigest         string `json:"observed_source_digest"`
	ObservedArtifactSourceDigest string `json:"observed_artifact_source_digest"`
	ObservedGeneratedJudgeDigest string `json:"observed_generated_judge_digest"`
	ObservedIndependentDigest    string `json:"observed_independent_digest"`
	UpperDecision                string `json:"upper_decision,omitempty"`
}

type revisionWireResult struct {
	CaseID           string   `json:"case_id"`
	Decision         string   `json:"decision"`
	MatchedCondition string   `json:"matched_condition"`
	Stage            string   `json:"stage"`
	Step             int      `json:"step"`
	Reason           string   `json:"reason"`
	UnknownClass     string   `json:"unknown_class"`
	NextOperation    string   `json:"next_operation"`
	BlockedBy        []string `json:"blocked_by"`
	PolicyDigest     string   `json:"policy_digest"`
	SemanticDigest   string   `json:"semantic_digest"`
	Denominator      int      `json:"fixed_denominator"`
}

type revisionWirePair struct {
	Baseline  revisionWireCase `json:"baseline"`
	Candidate revisionWireCase `json:"candidate"`
}

type revisionWireRequest struct {
	ExpectedSourceDigest string             `json:"expected_source_digest"`
	Condition            string             `json:"condition"`
	FromDecision         string             `json:"from_decision"`
	ToDecision           string             `json:"to_decision"`
	Cases                []revisionWirePair `json:"cases"`
}

type revisionWirePending struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type revisionWireExecution struct {
	GeneratedJudgeSource string               `json:"generated_judge_source"`
	GeneratedJudgeDigest string               `json:"generated_judge_digest"`
	DeclaredInputs       []revisionWireCase   `json:"declared_inputs"`
	SourceResults        []revisionWireResult `json:"source_results"`
	FirstResults         []revisionWireResult `json:"first_results"`
	ReplayResults        []revisionWireResult `json:"replay_results"`
	Complete             bool                 `json:"complete"`
	ExecutionError       string               `json:"execution_error,omitempty"`
	WallMilliseconds     int64                `json:"wall_ms"`
}

type revisionWireTransition struct {
	CaseID                      string `json:"case_id"`
	InputsIdentical             bool   `json:"inputs_identical"`
	BaselineCondition           string `json:"baseline_condition"`
	CandidateCondition          string `json:"candidate_condition"`
	BaselineDecision            string `json:"baseline_decision"`
	CandidateDecision           string `json:"candidate_decision"`
	RequestedTransitionObserved bool   `json:"requested_transition_observed"`
	CausalAttribution           string `json:"causal_attribution"`
}

type revisionWireCounts struct {
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

type revisionWireObservation struct {
	Schema                string                   `json:"schema"`
	SourceFile            string                   `json:"source_file"`
	RequestDigest         string                   `json:"canonical_request_digest"`
	RequestArtifactDigest string                   `json:"request_artifact_digest,omitempty"`
	Request               revisionWireRequest      `json:"request"`
	OriginalPolicy        revisionWirePolicy       `json:"original_policy"`
	CandidatePolicy       revisionWirePolicy       `json:"candidate_policy"`
	CandidateSource       string                   `json:"candidate_source"`
	ChangedCoordinates    []string                 `json:"changed_coordinates"`
	Baseline              revisionWireExecution    `json:"baseline"`
	Candidate             revisionWireExecution    `json:"candidate"`
	Transitions           []revisionWireTransition `json:"transitions"`
	Counts                revisionWireCounts       `json:"counts"`
	ExecutionStatus       string                   `json:"execution_status"`
	ExecutionConformance  string                   `json:"execution_conformance"`
	Admission             revisionWirePending      `json:"admission"`
	Pending               []revisionWirePending    `json:"pending"`
	InputProvenance       string                   `json:"input_provenance"`
	RepositoryObservation string                   `json:"repository_observation"`
	Improvement           string                   `json:"improvement"`
	MutationAuthority     int                      `json:"mutation_authority"`
	PromotionAuthority    int                      `json:"promotion_authority"`
}
