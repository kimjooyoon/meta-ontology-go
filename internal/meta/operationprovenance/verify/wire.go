package verify

type lineage struct {
	Producer  string `json:"producer"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Evidence  string `json:"evidence_path"`
}
type provenance struct {
	Source    string `json:"source_digest"`
	Semantic  string `json:"semantic_digest"`
	Producer  string `json:"producer"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Evidence  string `json:"evidence_path"`
	Mutation  string `json:"scenario_mutation"`
}
type issue struct {
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Cause     string   `json:"cause"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}
type claimTransition struct {
	Prior          string     `json:"prior_claim"`
	Next           string     `json:"next_claim"`
	Decision       string     `json:"conformance_decision"`
	Resolution     string     `json:"subject_resolution"`
	Transition     string     `json:"transition"`
	Stage          string     `json:"stage"`
	Step           string     `json:"step"`
	Reason         string     `json:"reason"`
	EvidenceDigest string     `json:"evidence_digest"`
	Provenance     provenance `json:"provenance"`
}
type metricResult struct {
	ID              string          `json:"id"`
	Family          string          `json:"family"`
	Claim           string          `json:"claim"`
	Numerator       int             `json:"numerator"`
	Denominator     int             `json:"denominator"`
	Decision        string          `json:"decision"`
	Resolution      string          `json:"subject_resolution"`
	EvaluationState string          `json:"evaluation_state"`
	Lineage         lineage         `json:"lineage"`
	Issue           *issue          `json:"issue,omitempty"`
	Transition      claimTransition `json:"claim_transition"`
}
