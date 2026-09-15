package languagesemantic

type SyntaxSummary struct {
	Satisfied    int `json:"satisfied"`
	Total        int `json:"total"`
	ValidCases   int `json:"valid_cases"`
	InvalidCases int `json:"invalid_cases"`
	GoooLines    int `json:"gooo_lines"`
}

type GoooFile struct {
	Path         string `json:"path"`
	GoooLines    int    `json:"gooo_lines"`
	SourceDigest string `json:"source_digest"`
}

type Source struct {
	ExpectedHeadSHA      string        `json:"expected_head_sha"`
	ConceptID            string        `json:"concept_id"`
	Producer             string        `json:"producer"`
	Consumer             string        `json:"consumer"`
	MetaOperation        string        `json:"meta_operation"`
	RegistryDigest       string        `json:"registry_digest"`
	SyntaxArtifactDigest string        `json:"syntax_artifact_digest"`
	SyntaxReportDigest   string        `json:"syntax_report_digest"`
	SyntaxSummary        SyntaxSummary `json:"syntax_summary"`
	GoooFiles            []GoooFile    `json:"gooo_files"`
	ObservationKnown     bool          `json:"observation_known"`
	ConceptBound         bool          `json:"concept_bound"`
}
