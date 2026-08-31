package languagegointeroperation

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
