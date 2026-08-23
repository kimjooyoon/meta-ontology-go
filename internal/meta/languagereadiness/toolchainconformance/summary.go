package toolchainconformance

type Summary struct {
	SurfacesSatisfied int `json:"surfaces_satisfied"`
	SurfacesTotal int `json:"surfaces_total"`
	ReadinessBPS int `json:"readiness_bps"`
	CasesSatisfied int `json:"cases_satisfied"`
	CasesTotal int `json:"cases_total"`
	ExecutedCases int `json:"executed_cases"`
	CaseReadinessBPS int `json:"case_readiness_bps"`
	IndicatorsSatisfied int `json:"indicators_satisfied"`
	IndicatorsTotal int `json:"indicators_total"`
	ProofsPassed int `json:"proofs_passed"`
	ProofsTotal int `json:"proofs_total"`
	ProofReadinessBPS int `json:"proof_readiness_bps"`
	HeadBindings int `json:"head_bindings"`
	TamperRejections int `json:"tamper_rejections"`
	TamperTotal int `json:"tamper_total"`
	ConceptBindings int `json:"concept_bindings"`
	CodeBindings int `json:"code_bindings"`
	MetricBindings int `json:"metric_bindings"`
	UseCaseBindings int `json:"use_case_bindings"`
	MissingSurfaces int `json:"missing_surfaces"`
	UnexpectedSurfaces int `json:"unexpected_surfaces"`
	SchemaMismatches int `json:"schema_mismatches"`
	HeadMismatches int `json:"head_mismatches"`
	DecisionMismatches int `json:"decision_mismatches"`
	ResolutionDescents int `json:"resolution_descents"`
	CaseMismatches int `json:"case_mismatches"`
	IndicatorFailures int `json:"indicator_failures"`
	ProofFailures int `json:"proof_failures"`
	Unresolved int `json:"unresolved"`
	DigestFailures int `json:"digest_failures"`
	RegistryDrift int `json:"registry_drift"`
	ConceptDrift int `json:"concept_drift"`
	RepositoryWrites int `json:"repository_writes"`
	MutationAuthorities int `json:"mutation_authorities"`
}
