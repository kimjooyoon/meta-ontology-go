package verify

type lineage struct {
	Producer  string `json:"producer"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Evidence  string `json:"evidence_path"`
}
type relationObservation struct {
	Relation         string `json:"relation"`
	DeclaredEndpoint string `json:"declared_endpoint"`
	ObservedEndpoint string `json:"observed_endpoint,omitempty"`
	ObservedArtifact string `json:"observed_artifact"`
	EvidenceKind     string `json:"evidence_kind"`
	ObservedDigest   string `json:"observed_digest"`
	RelationStatus   string `json:"relation_status"`
	Stage            string `json:"stage"`
	Step             string `json:"step"`
	Reason           string `json:"reason"`
	Cause            string `json:"cause"`
}
type provenance struct {
	Source    string            `json:"source_digest"`
	Semantic  string            `json:"semantic_digest"`
	Producer  string            `json:"producer"`
	Consumer  string            `json:"consumer"`
	Operation string            `json:"meta_operation"`
	Evidence  string            `json:"evidence_path"`
	Mutation  string            `json:"scenario_mutation"`
	Artifacts map[string]string `json:"artifact_digests"`
}
type issue struct {
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Cause     string   `json:"cause"`
	Detail    string   `json:"detail,omitempty"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}
type claimTransition struct {
	Proposition       string     `json:"proposition"`
	Prior             string     `json:"prior_claim"`
	Next              string     `json:"next_claim"`
	Decision          string     `json:"conformance_decision"`
	SourceResolution  string     `json:"source_resolution"`
	LineageResolution string     `json:"lineage_resolution"`
	Transition        string     `json:"transition"`
	Stage             string     `json:"stage"`
	Step              string     `json:"step"`
	Reason            string     `json:"reason"`
	EvidenceDigest    string     `json:"evidence_digest"`
	Provenance        provenance `json:"provenance"`
}
type metricResult struct {
	ID                string                `json:"id"`
	Family            string                `json:"family"`
	Claim             string                `json:"claim"`
	Numerator         int                   `json:"numerator"`
	Denominator       int                   `json:"denominator"`
	Decision          string                `json:"decision"`
	Proposition       string                `json:"proposition"`
	SourceResolution  string                `json:"source_resolution"`
	LineageResolution string                `json:"lineage_resolution"`
	EvaluationState   string                `json:"evaluation_state"`
	Lineage           lineage               `json:"lineage"`
	Relations         []relationObservation `json:"relations"`
	Issue             *issue                `json:"issue,omitempty"`
	Transition        claimTransition       `json:"claim_transition"`
}
