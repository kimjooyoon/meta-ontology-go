package languageartifactoracle

type CaseResult struct {
	ID               string        `json:"id"`
	Status           string        `json:"status"`
	ExpectedDecision string        `json:"expected_decision"`
	ExpectedReason   string        `json:"expected_reason"`
	ObservedDecision string        `json:"observed_decision"`
	ObservedResolution string      `json:"observed_resolution"`
	ObservedReason   string        `json:"observed_reason"`
	ProofChoice      string        `json:"proof_choice"`
	MetaOperation    string        `json:"meta_operation"`
	Coordinate       Coordinate    `json:"coordinate"`
	Checks           []CheckResult `json:"checks"`
	SourceDigest     string        `json:"source_digest"`
	ArtifactDigest   string        `json:"artifact_digest"`
}

type Summary struct {
	CasesSatisfied                    int `json:"cases_satisfied"`
	CasesTotal                        int `json:"cases_total"`
	ExactSourceBindings               int `json:"exact_source_bindings"`
	ResealedForgeriesRejected         int `json:"resealed_forgeries_rejected"`
	UnknownDecisionsRejected          int `json:"unknown_decisions_rejected"`
	LowerResolutions                  int `json:"lower_resolutions"`
	LegacyValidatorCounterexamples    int `json:"legacy_validator_counterexamples"`
	ProducerDependencies              int `json:"producer_dependencies"`
	UnknownChecks                     int `json:"unknown_checks"`
	SemanticCorrectnessClaims         int `json:"semantic_correctness_claims"`
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

type Report struct {
	Schema             string       `json:"schema"`
	Scope              string       `json:"scope"`
	HeadSHA            string       `json:"head_sha"`
	Decision           string       `json:"decision"`
	Resolution         string       `json:"resolution"`
	Reason             string       `json:"reason"`
	ContractDigest     string       `json:"contract_digest"`
	IndependenceDigest string       `json:"independence_digest"`
	LegacyDigest       string       `json:"legacy_digest"`
	Cases              []CaseResult `json:"cases"`
	Summary            Summary      `json:"summary"`
	Indicators         []Indicator  `json:"indicators"`
	NotClaimed         []string     `json:"not_claimed"`
	RepositoryWrites   int          `json:"repository_writes"`
	MutationAuthority  bool         `json:"mutation_authority"`
	Digest             string       `json:"digest"`
}
