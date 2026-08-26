package languagediagnosticprovenance

type Source struct {
	ExpectedHeadSHA       string `json:"expected_head_sha"`
	ConceptID             string `json:"concept_id"`
	Producer              string `json:"producer"`
	Consumer              string `json:"consumer"`
	MetaOperation         string `json:"meta_operation"`
	ConceptArtifactDigest string `json:"concept_artifact_digest"`
	CatalogDigest         string `json:"catalog_digest"`
	RegistryDigest        string `json:"registry_digest"`
	Toolchain             string `json:"toolchain"`
	TokenReference        string `json:"token_reference"`
	ScannerReference      string `json:"scanner_reference"`
	TypesReference        string `json:"types_reference"`
	LineDirective         string `json:"line_directive_reference"`
	MacroReference        string `json:"macro_reference"`
}

type Indicator struct {
	MetricID      string     `json:"metric_id"`
	Class         string     `json:"class"`
	ProofChoice   string     `json:"proof_choice"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	MetaOperation string     `json:"meta_operation"`
	Resolution    Resolution `json:"resolution"`
	Value         int        `json:"value"`
	Target        int        `json:"target"`
	Satisfied     bool       `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Report struct {
	Schema             string        `json:"schema"`
	Decision           Decision      `json:"decision"`
	Resolution         Resolution    `json:"resolution"`
	ReasonCode         string        `json:"reason_code"`
	Source             Source        `json:"source"`
	Summary            Summary       `json:"summary"`
	Cases              []CaseResult  `json:"cases"`
	Stages             []StepReceipt `json:"stages"`
	Indicators         []Indicator   `json:"indicators"`
	Proofs             []Proof       `json:"proofs"`
	RepositoryWrites   int           `json:"repository_writes"`
	MutationAuthorized bool          `json:"mutation_authorized"`
	ReportDigest       string        `json:"report_digest"`
}
