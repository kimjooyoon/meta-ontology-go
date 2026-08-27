package operationprovenance

type Lineage struct {
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	EvidencePath  string `json:"evidence_path"`
}

type RelationObservation struct {
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

type Provenance struct {
	SourceDigest     string            `json:"source_digest"`
	SemanticDigest   string            `json:"semantic_digest"`
	Producer         string            `json:"producer"`
	Consumer         string            `json:"consumer"`
	MetaOperation    string            `json:"meta_operation"`
	EvidencePath     string            `json:"evidence_path"`
	ScenarioMutation string            `json:"scenario_mutation"`
	ArtifactDigests  map[string]string `json:"artifact_digests"`
}

type Issue struct {
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Detail    string   `json:"detail,omitempty"`
	Cause     string   `json:"cause"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}

type ClaimTransition struct {
	Proposition         string     `json:"proposition"`
	PriorClaim          string     `json:"prior_claim"`
	NextClaim           string     `json:"next_claim"`
	ConformanceDecision string     `json:"conformance_decision"`
	SourceResolution    string     `json:"source_resolution"`
	LineageResolution   string     `json:"lineage_resolution"`
	Transition          string     `json:"transition"`
	Stage               string     `json:"stage"`
	Step                string     `json:"step"`
	Reason              string     `json:"reason"`
	EvidenceDigest      string     `json:"evidence_digest"`
	Provenance          Provenance `json:"provenance"`
}
