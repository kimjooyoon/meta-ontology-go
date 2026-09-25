package languagediagnosticprovenancebinding

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Artifact struct {
	Schema             string       `json:"schema"`
	Decision           string       `json:"decision"`
	Resolution         string       `json:"resolution"`
	ReasonCode         string       `json:"reason_code"`
	ExpectedHeadSHA    string       `json:"expected_head_sha"`
	ConceptDigest      string       `json:"concept_digest"`
	ReadinessDigest    string       `json:"readiness_digest"`
	ProvenanceDigest   string       `json:"provenance_digest"`
	Coordinates        []Coordinate `json:"coordinates"`
	Summary            Summary      `json:"summary"`
	Indicators         []Indicator  `json:"indicators"`
	Proofs             []Proof      `json:"proofs"`
	RepositoryWrites   int          `json:"repository_writes"`
	MutationAuthorized bool         `json:"mutation_authorized"`
	ArtifactDigest     string       `json:"artifact_digest"`
}
