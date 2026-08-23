package languagesemanticbinding

type semanticArtifact struct {
	Schema             string              `json:"schema"`
	Decision           string              `json:"decision"`
	ReasonCode         string              `json:"reason_code"`
	Resolution         string              `json:"resolution"`
	ReportDigest       string              `json:"report_digest"`
	RepositoryWrites   int                 `json:"repository_writes"`
	MutationAuthorized bool                `json:"mutation_authorized"`
	Source             semanticSource      `json:"source"`
	Summary            semanticSummary     `json:"summary"`
	Cases              []semanticCase      `json:"cases"`
	Indicators         []semanticIndicator `json:"indicators"`
	Proofs             []semanticProof     `json:"proofs"`
}

type semanticSummary struct {
	Satisfied                 int `json:"satisfied"`
	Total                     int `json:"total"`
	Executed                  int `json:"executed"`
	NotSatisfied              int `json:"not_satisfied"`
	Unresolved                int `json:"unresolved"`
	ReadinessBPS              int `json:"readiness_bps"`
	SourceModels              int `json:"source_models"`
	NormalizedIRs             int `json:"normalized_irs"`
	SemanticReplays           int `json:"semantic_replays"`
	ProvenanceReplays         int `json:"provenance_replays"`
	EvidenceReplays           int `json:"evidence_replays"`
	PresentationLaws          int `json:"presentation_laws"`
	CandidateAuthorityLaws    int `json:"candidate_authority_laws"`
	DeterministicAuthorityLaws int `json:"deterministic_authority_laws"`
	UpstreamRejections        int `json:"upstream_rejections"`
	UnregisteredGooo          int `json:"unregistered_gooo"`
	MissingRegistered         int `json:"missing_registered"`
	StageOrderViolations      int `json:"stage_order_violations"`
	EffectfulStages           int `json:"effectful_stages"`
	RegistryDrift             int `json:"registry_drift"`
}
