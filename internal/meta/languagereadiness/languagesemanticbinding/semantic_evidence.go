package languagesemanticbinding

type semanticSource struct {
	ExpectedHeadSHA  string                `json:"expected_head_sha"`
	ConceptID        string                `json:"concept_id"`
	Producer         string                `json:"producer"`
	Consumer         string                `json:"consumer"`
	MetaOperation    string                `json:"meta_operation"`
	RegistryDigest   string                `json:"registry_digest"`
	SyntaxSummary    semanticSyntaxSummary `json:"syntax_summary"`
	ObservationKnown bool                  `json:"observation_known"`
	ConceptBound     bool                  `json:"concept_bound"`
}

type semanticSyntaxSummary struct {
	Satisfied    int `json:"satisfied"`
	Total        int `json:"total"`
	ValidCases   int `json:"valid_cases"`
	InvalidCases int `json:"invalid_cases"`
	GoooLines    int `json:"gooo_lines"`
}

type semanticCase struct {
	Status string `json:"status"`
}

type semanticIndicator struct {
	MetricID  string `json:"metric_id"`
	Class     string `json:"class"`
	Value     int    `json:"value"`
	Target    int    `json:"target"`
	Satisfied bool   `json:"satisfied"`
}

type semanticProof struct {
	Choice         string `json:"choice"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}
