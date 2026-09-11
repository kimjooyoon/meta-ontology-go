package selfimprovementcandidate

type Candidate struct {
	Schema                      string     `json:"schema"`
	ID                          string     `json:"id"`
	SourceObservationDigest     string     `json:"source_observation_digest"`
	GapID                       string     `json:"gap_id"`
	SourceNonClaim              string     `json:"source_nonclaim"`
	ExperimentKind              string     `json:"experiment_kind"`
	Before                      Coordinate `json:"before"`
	Target                      Coordinate `json:"target"`
	ProofChoice                 string     `json:"proof_choice"`
	MetaOperation               string     `json:"meta_operation"`
	ExecutionInputDigest        string     `json:"execution_input_digest"`
	ExecutionAuthorized         bool       `json:"execution_authorized"`
	MutationAuthorized          bool       `json:"mutation_authorized"`
	PromotionAuthorized         bool       `json:"promotion_authorized"`
	AutomaticAdoptionAuthorized bool       `json:"automatic_adoption_authorized"`
	Digest                      string     `json:"digest"`
}

type gapPolicy struct {
	SourceNonClaim string
	GapID          string
	ExperimentKind string
	ProofChoice    string
	MetaOperation  string
}

var gapPolicies = []gapPolicy{{
	SourceNonClaim: "value-level computation", GapID: "value-level-computation",
	ExperimentKind: "VALUE_WITNESS_EXPERIMENT", ProofChoice: "COHERENCE",
	MetaOperation: "propose-value-level-witness-experiment",
}}

func buildCandidate(sourceDigest string, policy gapPolicy) Candidate {
	candidate := Candidate{Schema: CandidateSchema, SourceObservationDigest: sourceDigest,
		GapID: policy.GapID, SourceNonClaim: policy.SourceNonClaim,
		ExperimentKind: policy.ExperimentKind, Before: coordinate(0, 1), Target: coordinate(1, 1),
		ProofChoice: policy.ProofChoice, MetaOperation: policy.MetaOperation,
		ID: digestJSON(struct{ Source, Policy string }{sourceDigest, PolicyVersion + ":" + policy.GapID})}
	candidate.Digest = candidateDigest(candidate)
	return candidate
}

func candidateDigest(candidate Candidate) string {
	candidate.Digest = ""
	return digestJSON(candidate)
}
