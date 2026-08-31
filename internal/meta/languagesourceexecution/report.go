package languagesourceexecution

const ArtifactSchema = "gooo/language-source-execution-artifact/v1"

type CaseResult struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	ProofChoice    string `json:"proof_choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Summary struct {
	CasesSatisfied       int `json:"cases_satisfied"`
	CasesTotal           int `json:"cases_total"`
	SourceExecutions     int `json:"source_executions"`
	DeterministicReplays int `json:"deterministic_replays"`
	DiagnosticRejections int `json:"diagnostic_rejections"`
	ExecutionEvents      int `json:"execution_events"`
	Unknowns             int `json:"unknowns"`
	NotSatisfied         int `json:"not_satisfied"`
	RepositoryWrites     int `json:"repository_writes"`
	MutationAuthorities  int `json:"mutation_authorities"`
}

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
	Schema            string       `json:"schema"`
	HeadSHA           string       `json:"head_sha"`
	Decision          string       `json:"decision"`
	Resolution        string       `json:"resolution"`
	Reason            string       `json:"reason"`
	ContractDigest    string       `json:"contract_digest"`
	Cases             []CaseResult `json:"cases"`
	Summary           Summary      `json:"summary"`
	Indicators        []Indicator  `json:"indicators"`
	Proofs            []Proof      `json:"proofs"`
	RepositoryWrites  int          `json:"repository_writes"`
	MutationAuthority bool         `json:"mutation_authority"`
	NotClaimed        []string     `json:"not_claimed"`
	Digest            string       `json:"digest"`
}
