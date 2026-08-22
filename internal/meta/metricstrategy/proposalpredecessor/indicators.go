package proposalpredecessor

import artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"

type Summary struct {
	ObservedRuns         int `json:"observed_runs"`
	ExactRuns            int `json:"exact_runs"`
	ObservedArtifacts    int `json:"observed_artifacts"`
	ExactArtifacts       int `json:"exact_artifacts"`
	ValidCandidates      int `json:"valid_candidates"`
	AmbiguousCandidates  int `json:"ambiguous_candidates"`
	UnresolvedCandidates int `json:"unresolved_candidates"`
	RepositoryWrites     int `json:"repository_writes"`
	SelectionBPS         int `json:"selection_bps"`
	ProofsPassed         int `json:"proofs_passed"`
	ProofsTotal          int `json:"proofs_total"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	ID             string `json:"id"`
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	Passed         bool   `json:"passed"`
	EvidenceDigest string `json:"evidence_digest"`
}

func buildIndicators(summary Summary) []Indicator {
	return []Indicator{
		{MetricID: "gooo.metric.language.change-proposal-predecessor-selection-bps.v1", Class: "outcome", ProofChoice: "COHERENCE", MetaOperation: "select-merged-change-proposal", Value: summary.SelectionBPS, Target: 10000, Satisfied: summary.SelectionBPS == 10000},
		{MetricID: "gooo.metric.language.change-proposal-predecessor-valid-candidates.v1", Class: "driver", ProofChoice: "FOUNDATION", MetaOperation: "count-valid-proposal-predecessors", Value: summary.ValidCandidates, Target: 1, Satisfied: summary.ValidCandidates == 1},
		{MetricID: "gooo.metric.language.change-proposal-predecessor-ambiguity.guardrail.v1", Class: "guardrail", ProofChoice: "REGRESSION", MetaOperation: "reject-ambiguous-proposal-predecessor", Value: summary.AmbiguousCandidates, Target: 0, Satisfied: summary.AmbiguousCandidates == 0},
		{MetricID: "gooo.metric.language.change-proposal-predecessor-unresolved.guardrail.v1", Class: "guardrail", ProofChoice: "FOUNDATION", MetaOperation: "lower-resolution-on-unknown-proposal-predecessor", Value: summary.UnresolvedCandidates, Target: 0, Satisfied: summary.UnresolvedCandidates == 0},
		{MetricID: "gooo.metric.language.change-proposal-predecessor-writes.guardrail.v1", Class: "guardrail", ProofChoice: "FOUNDATION", MetaOperation: "preserve-read-only-proposal-selection", Value: summary.RepositoryWrites, Target: 0, Satisfied: summary.RepositoryWrites == 0},
	}
}

func makeProof(id, choice, operation string, passed bool, evidence any) (Proof, error) {
	digest, err := artifact.Digest(evidence)
	return Proof{ID: id, Choice: choice, MetaOperation: operation, Passed: passed, EvidenceDigest: digest}, err
}
