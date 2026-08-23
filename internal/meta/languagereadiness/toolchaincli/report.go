package toolchaincli

type Source struct {
	ExpectedHeadSHA  string `json:"expected_head_sha"`
	GoVersion        string `json:"go_version"`
	ConceptDigest    string `json:"concept_digest"`
	RegistryDigest   string `json:"registry_digest"`
	ExecutableDigest string `json:"executable_digest"`
	ObservationKnown bool   `json:"observation_known"`
}

type Report struct {
	Schema             string       `json:"schema"`
	Decision           Decision     `json:"decision"`
	Resolution         Resolution   `json:"resolution"`
	ReasonCode         string       `json:"reason_code"`
	Source             Source       `json:"source"`
	Summary            Summary      `json:"summary"`
	Cases              []CaseResult `json:"cases"`
	Indicators         []Indicator  `json:"indicators"`
	Proofs             []Proof      `json:"proofs"`
	Stages             []Stage      `json:"stages"`
	RepositoryWrites   int          `json:"repository_writes"`
	MutationAuthorized bool         `json:"mutation_authorized"`
	ReportDigest       string       `json:"report_digest"`
}
