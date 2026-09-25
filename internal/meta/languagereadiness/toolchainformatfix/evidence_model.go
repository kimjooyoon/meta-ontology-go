package toolchainformatfix

type Indicator struct {
	MetricID      string     `json:"metric_id"`
	Class         string     `json:"class"`
	ProofChoice   string     `json:"proof_choice"`
	MetaOperation string     `json:"meta_operation"`
	Value         int        `json:"value"`
	Target        int        `json:"target"`
	Resolution    Resolution `json:"resolution"`
	Satisfied     bool       `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
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
	RepositoryWrites   int          `json:"repository_writes"`
	MutationAuthorized bool         `json:"mutation_authorized"`
	ReportDigest       string       `json:"report_digest"`
}
