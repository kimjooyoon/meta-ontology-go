package toolchainusecases

const (
	RegistrySchema  = "gooo/toolchain-executable-use-case-registry/v1"
	ReportSchema    = "gooo/toolchain-executable-use-cases/v1"
	DecisionPass    = "PASS"
	DecisionClosed  = "FAIL_CLOSED"
	ResolutionExact = "EXACT"
	ResolutionLower = "LOWER_RESOLUTION"
)

const (
	caseCanonical = "canonical-concept-artifact"
	caseReplay    = "replay-digest-tamper"
	caseWrite     = "repository-write-tamper"
	totalCases    = 3
)

func expectedRegistry() Registry {
	return Registry{Schema: RegistrySchema, Cases: []CaseDefinition{
		{ID: caseCanonical, ProofChoice: "FOUNDATION", Mutation: "NONE",
			ExpectedDecision: DecisionPass, MetaOperation: "consume-canonical-concept-artifact"},
		{ID: caseReplay, ProofChoice: "COHERENCE", Mutation: "REPLAY_DIGEST",
			ExpectedDecision: DecisionClosed, MetaOperation: "reject-replay-digest-tamper"},
		{ID: caseWrite, ProofChoice: "REGRESSION", Mutation: "REPOSITORY_WRITE",
			ExpectedDecision: DecisionClosed, MetaOperation: "reject-repository-write-tamper"},
	}}
}

type Registry struct {
	Schema string           `json:"schema"`
	Cases  []CaseDefinition `json:"cases"`
}

type CaseDefinition struct {
	ID               string `json:"id"`
	ProofChoice      string `json:"proof_choice"`
	Mutation         string `json:"mutation"`
	ExpectedDecision string `json:"expected_decision"`
	MetaOperation    string `json:"meta_operation"`
}

type Source struct {
	ExpectedHeadSHA         string `json:"expected_head_sha"`
	ConceptArtifactDigest   string `json:"concept_artifact_digest"`
	CatalogDigest           string `json:"catalog_digest"`
	RegistryDigest          string `json:"registry_digest"`
	ConceptRepositoryWrites int    `json:"concept_repository_writes"`
}

type CaseResult struct {
	Definition       CaseDefinition `json:"definition"`
	ObservedDecision string         `json:"observed_decision"`
	Status           string         `json:"status"`
	EvidenceDigest   string         `json:"evidence_digest"`
}

type Summary struct {
	Satisfied        int `json:"satisfied"`
	Total            int `json:"total"`
	Executed         int `json:"executed"`
	NotSatisfied     int `json:"not_satisfied"`
	Unresolved       int `json:"unresolved"`
	ReadinessBPS     int `json:"readiness_bps"`
	PassPaths        int `json:"pass_paths"`
	FailClosedPaths  int `json:"fail_closed_paths"`
	RepositoryWrites int `json:"repository_writes"`
}
