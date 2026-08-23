package languagegointeroperation

type Evidence struct {
	ActualOutcome       string `json:"actual_outcome"`
	FailureStage        string `json:"failure_stage,omitempty"`
	ErrorCode           string `json:"error_code,omitempty"`
	SourceDigest        string `json:"source_digest,omitempty"`
	ReplaySourceDigest  string `json:"replay_source_digest,omitempty"`
	CanonicalDigest     string `json:"canonical_digest,omitempty"`
	ReplayCanonical     string `json:"replay_canonical_digest,omitempty"`
	APIDigest           string `json:"api_digest,omitempty"`
	ReplayAPIDigest     string `json:"replay_api_digest,omitempty"`
	SourceMapDigest     string `json:"source_map_digest,omitempty"`
	ReplaySourceMap     string `json:"replay_source_map_digest,omitempty"`
	SourceMapMappings   int    `json:"source_map_mappings"`
	ExportedObjects     int    `json:"exported_objects"`
	GenericMethods      int    `json:"generic_methods"`
	AliasNodes          int    `json:"alias_nodes"`
	ASTReifications     int    `json:"ast_reifications"`
	CanonicalReplay     bool   `json:"canonical_replay"`
	TypeIdentityReplay  bool   `json:"type_identity_replay"`
	Rejected            bool   `json:"rejected"`
	InvalidAccepted     bool   `json:"invalid_accepted"`
	UnknownAccepted     bool   `json:"unknown_accepted"`
	ImportAccepted      bool   `json:"import_accepted"`
	Effects             int    `json:"effects"`
}

type CaseResult struct {
	Definition Definition `json:"definition"`
	Evidence   Evidence   `json:"evidence"`
	Status     CaseStatus `json:"status"`
	Digest     string     `json:"evidence_digest"`
}

type Summary struct {
	Satisfied            int `json:"satisfied"`
	Total                int `json:"total"`
	Executed             int `json:"executed"`
	NotSatisfied         int `json:"not_satisfied"`
	Unresolved           int `json:"unresolved"`
	ReadinessBPS         int `json:"readiness_bps"`
	GeneratorProjections int `json:"generator_projections"`
	Go127Boundaries      int `json:"go_1_27_boundaries"`
	GuardrailRejections  int `json:"guardrail_rejections"`
	PositiveAccepted     int `json:"positive_accepted"`
	CanonicalReplays     int `json:"canonical_replays"`
	TypeIdentityReplays  int `json:"type_identity_replays"`
	SourceMaps           int `json:"source_maps"`
	GenericMethods       int `json:"generic_methods"`
	AliasNodes           int `json:"alias_nodes"`
	ASTReifications      int `json:"ast_reifications"`
	ConceptBindings      int `json:"concept_bindings"`
	CodeBindings         int `json:"code_bindings"`
	MetricBindings       int `json:"metric_bindings"`
	UseCaseBindings      int `json:"use_case_bindings"`
	RegistryDrift        int `json:"registry_drift"`
	ToolchainMatches     int `json:"toolchain_matches"`
	InvalidAcceptances   int `json:"invalid_acceptances"`
	UnknownAcceptances   int `json:"unknown_acceptances"`
	ImportAcceptances    int `json:"import_acceptances"`
	EffectfulStages      int `json:"effectful_stages"`
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
	Toolchain             string `json:"toolchain"`
	GoReleaseNotes        string `json:"go_release_notes"`
	MacroReference        string `json:"macro_reference"`
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
