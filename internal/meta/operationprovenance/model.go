package operationprovenance

type Lineage struct {
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	EvidencePath  string `json:"evidence_path"`
}

type Provenance struct {
	SourceDigest     string `json:"source_digest"`
	SemanticDigest   string `json:"semantic_digest"`
	Producer         string `json:"producer"`
	Consumer         string `json:"consumer"`
	MetaOperation    string `json:"meta_operation"`
	EvidencePath     string `json:"evidence_path"`
	ScenarioMutation string `json:"scenario_mutation"`
}

type Issue struct {
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Cause     string   `json:"cause"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}

type ClaimTransition struct {
	PriorClaim          string     `json:"prior_claim"`
	NextClaim           string     `json:"next_claim"`
	ConformanceDecision string     `json:"conformance_decision"`
	SubjectResolution   string     `json:"subject_resolution"`
	Transition          string     `json:"transition"`
	Stage               string     `json:"stage"`
	Step                string     `json:"step"`
	Reason              string     `json:"reason"`
	EvidenceDigest      string     `json:"evidence_digest"`
	Provenance          Provenance `json:"provenance"`
}
