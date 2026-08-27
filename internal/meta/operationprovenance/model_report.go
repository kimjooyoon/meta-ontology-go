package operationprovenance

type ImportCheck struct {
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	Status      string `json:"status"`
}

type Report struct {
	Schema                  string               `json:"schema"`
	Status                  string               `json:"status"`
	ConformanceDecision     string               `json:"conformance_decision"`
	SubjectResolution       string               `json:"subject_resolution"`
	SourceDigest            string               `json:"source_digest,omitempty"`
	CanonicalSemanticDigest string               `json:"canonical_semantic_digest,omitempty"`
	ReceiptDigest           string               `json:"receipt_digest,omitempty"`
	ScenarioCount           int                  `json:"scenario_count"`
	MetricCount             int                  `json:"metric_count"`
	FailClosedCount         int                  `json:"fail_closed_count"`
	DirectUnknowns          int                  `json:"direct_unknowns"`
	DependencyBlocks        int                  `json:"dependency_blocks"`
	TransitionCounts        map[string]int       `json:"transition_counts"`
	SourceReconstruction    SourceReconstruction `json:"source_reconstruction"`
	ProducerImport          ImportCheck          `json:"producer_import"`
	Issue                   *Issue               `json:"issue,omitempty"`
	Digest                  string               `json:"digest"`
}
