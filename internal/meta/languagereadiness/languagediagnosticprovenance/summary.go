package languagediagnosticprovenance

type Summary struct {
	Satisfied           int `json:"satisfied"`
	Total               int `json:"total"`
	Executed            int `json:"executed"`
	NotSatisfied        int `json:"not_satisfied"`
	Unresolved          int `json:"unresolved"`
	ReadinessBPS        int `json:"readiness_bps"`
	Traced              int `json:"traced"`
	GuardrailRejections int `json:"guardrail_rejections"`
	PhysicalPositions   int `json:"physical_positions"`
	LogicalPositions    int `json:"logical_positions"`
	SemanticBindings    int `json:"semantic_bindings"`
	LSPProjections      int `json:"lsp_projections"`
	CanonicalReplays    int `json:"canonical_replays"`
	OrderedDiagnostics  int `json:"ordered_diagnostics"`
	LineDirectiveRemaps int `json:"line_directive_remaps"`
	TypeClassifications int `json:"type_classifications"`
	ProvenanceSteps     int `json:"provenance_steps"`
	ConceptBindings     int `json:"concept_bindings"`
	CodeBindings        int `json:"code_bindings"`
	MetricBindings      int `json:"metric_bindings"`
	UseCaseBindings     int `json:"use_case_bindings"`
	RegistryDrift       int `json:"registry_drift"`
	ConceptDrift        int `json:"concept_drift"`
	ToolchainMatches    int `json:"toolchain_matches"`
	UnknownAcceptances  int `json:"unknown_acceptances"`
	MissingMapAccepts   int `json:"missing_map_acceptances"`
	AmbiguousAccepts    int `json:"ambiguous_map_acceptances"`
	InvalidAcceptances  int `json:"invalid_acceptances"`
	EffectfulStages     int `json:"effectful_stages"`
}
