package toolchainconformance

type Source struct {
	ExpectedHeadSHA       string `json:"expected_head_sha"`
	RegistryDigest        string `json:"registry_digest"`
	ConceptArtifactDigest string `json:"concept_artifact_digest"`
	CatalogDigest         string `json:"catalog_digest"`
	ObservationKnown      bool   `json:"observation_known"`
}

type SurfaceResult struct {
	ID             string `json:"id"`
	Schema         string `json:"schema"`
	Status         string `json:"status"`
	HeadSHA        string `json:"head_sha"`
	Cases          int    `json:"cases"`
	Indicators     int    `json:"indicators"`
	Proofs         int    `json:"proofs"`
	EvidenceDigest string `json:"evidence_digest"`
}

type CaseResult struct {
	ID               string `json:"id"`
	Mutation         string `json:"mutation"`
	Target           string `json:"target"`
	ExpectedDecision string `json:"expected_decision"`
	ObservedDecision string `json:"observed_decision"`
	Status           string `json:"status"`
	EvidenceDigest   string `json:"evidence_digest"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Resolution    string `json:"resolution"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Relation      string `json:"relation"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Report struct {
	Schema             string          `json:"schema"`
	Decision           string          `json:"decision"`
	Resolution         string          `json:"resolution"`
	ReasonCode         string          `json:"reason_code"`
	Source             Source          `json:"source"`
	Summary            Summary         `json:"summary"`
	Surfaces           []SurfaceResult `json:"surfaces"`
	Cases              []CaseResult    `json:"cases"`
	Indicators         []Indicator     `json:"indicators"`
	Proofs             []Proof         `json:"proofs"`
	RepositoryWrites   int             `json:"repository_writes"`
	MutationAuthorized bool            `json:"mutation_authorized"`
	ReportDigest       string          `json:"report_digest"`
}
