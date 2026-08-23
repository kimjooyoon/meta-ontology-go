package languagedeterministicquery

type Evidence struct {
	RequestDigest     string `json:"request_digest,omitempty"`
	ReplayRequest     string `json:"replay_request_digest,omitempty"`
	ResultDigest      string `json:"result_digest,omitempty"`
	ReplayResult      string `json:"replay_result_digest,omitempty"`
	PermutationResult string `json:"permutation_result_digest,omitempty"`
	GraphBefore       string `json:"graph_before,omitempty"`
	GraphAfter        string `json:"graph_after,omitempty"`
	DeterministicRows int    `json:"deterministic_rows"`
	CandidateRows     int    `json:"candidate_rows"`
	Rejected          bool   `json:"rejected"`
	CandidatePromoted bool   `json:"candidate_promoted"`
	UnknownAccepted   bool   `json:"unknown_accepted"`
	GraphMutated      bool   `json:"graph_mutated"`
	Error             string `json:"error,omitempty"`
}

type CaseResult struct {
	Definition Definition `json:"definition"`
	Evidence   Evidence   `json:"evidence"`
	Status     CaseStatus `json:"status"`
	Digest     string     `json:"evidence_digest"`
}

type Summary struct {
	Satisfied           int `json:"satisfied"`
	Total               int `json:"total"`
	Executed            int `json:"executed"`
	NotSatisfied        int `json:"not_satisfied"`
	Unresolved          int `json:"unresolved"`
	ReadinessBPS        int `json:"readiness_bps"`
	BindingPlans        int `json:"binding_plans"`
	LawPlans            int `json:"law_plans"`
	CanonicalReplays    int `json:"canonical_replays"`
	PermutationReplays  int `json:"permutation_replays"`
	ConceptBindings     int `json:"concept_bindings"`
	CodeBindings        int `json:"code_bindings"`
	MetricBindings      int `json:"metric_bindings"`
	UseCaseBindings     int `json:"use_case_bindings"`
	RegistryDrift       int `json:"registry_drift"`
	CandidatePromotions int `json:"candidate_promotions"`
	UnknownAcceptances  int `json:"unknown_acceptances"`
	GraphMutations      int `json:"graph_mutations"`
	EffectfulStages     int `json:"effectful_stages"`
}

type Source struct {
	ExpectedHeadSHA       string `json:"expected_head_sha"`
	ConceptID             string `json:"concept_id"`
	Producer              string `json:"producer"`
	Consumer              string `json:"consumer"`
	MetaOperation         string `json:"meta_operation"`
	ConceptArtifactDigest string `json:"concept_artifact_digest"`
	CatalogDigest         string `json:"catalog_digest"`
	RegistryDigest        string `json:"registry_digest"`
	ConceptBound          bool   `json:"concept_bound"`
}

type StageReceipt struct {
	Ordinal       int    `json:"ordinal"`
	Stage         string `json:"stage"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Status        string `json:"status"`
	Effects       int    `json:"effects"`
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
	Schema             string         `json:"schema"`
	Decision           Decision       `json:"decision"`
	Resolution         Resolution     `json:"resolution"`
	ReasonCode         string         `json:"reason_code"`
	Source             Source         `json:"source"`
	Summary            Summary        `json:"summary"`
	Cases              []CaseResult   `json:"cases"`
	Stages             []StageReceipt `json:"stages"`
	Indicators         []Indicator    `json:"indicators"`
	Proofs             []Proof        `json:"proofs"`
	RepositoryWrites   int            `json:"repository_writes"`
	MutationAuthorized bool           `json:"mutation_authorized"`
	ReportDigest       string         `json:"report_digest"`
}
