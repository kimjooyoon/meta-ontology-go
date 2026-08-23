package toolchainrelease

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
	ProofChoice    string `json:"proof_choice"`
	Claim          string `json:"claim"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Report struct {
	Schema           string       `json:"schema"`
	Decision         string       `json:"decision"`
	Reason           string       `json:"reason"`
	Resolution       string       `json:"resolution"`
	HeadSHA          string       `json:"head_sha"`
	CorpusDigest     string       `json:"corpus_digest"`
	ConceptDigest    string       `json:"concept_digest"`
	Summary          Summary      `json:"summary"`
	Cases            []CaseResult `json:"cases"`
	Indicators       []Indicator  `json:"indicators"`
	Proofs           []Proof      `json:"proofs"`
	RepositoryWrites int          `json:"repository_writes"`
	ReportDigest     string       `json:"report_digest"`
}
