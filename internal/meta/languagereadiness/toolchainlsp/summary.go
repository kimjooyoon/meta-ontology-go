package toolchainlsp

type Summary struct {
	CasesSatisfied          int `json:"cases_satisfied"`
	CasesTotal              int `json:"cases_total"`
	ReadinessBPS            int `json:"readiness_bps"`
	ProtocolCases           int `json:"protocol_cases"`
	CouplingCases           int `json:"coupling_cases"`
	AdvertisedCapabilities  int `json:"advertised_capabilities"`
	ReadFeatures            int `json:"read_features"`
	DiagnosticPaths         int `json:"diagnostic_paths"`
	NavigationPaths         int `json:"navigation_paths"`
	SymbolPaths             int `json:"symbol_paths"`
	SemanticTokenPaths      int `json:"semantic_token_paths"`
	UTF16Replays            int `json:"utf16_replays"`
	TranscriptReplays       int `json:"transcript_replays"`
	FailClosedPaths         int `json:"fail_closed_paths"`
	ConceptBindings         int `json:"concept_bindings"`
	CodeBindings            int `json:"code_bindings"`
	MetricBindings          int `json:"metric_bindings"`
	UseCaseBindings         int `json:"use_case_bindings"`
	MissingCases            int `json:"missing_cases"`
	UnexpectedCases         int `json:"unexpected_cases"`
	CaseFailures            int `json:"case_failures"`
	CapabilityGaps          int `json:"capability_gaps"`
	UnexpectedProtocolErrors int `json:"unexpected_protocol_errors"`
	DiagnosticGaps          int `json:"diagnostic_gaps"`
	NonstandardWireFields   int `json:"nonstandard_wire_fields"`
	StaleNavigationLeaks    int `json:"stale_navigation_leaks"`
	UnknownNavigationLeaks  int `json:"unknown_navigation_leaks"`
	FailClosedNavigationLeaks int `json:"fail_closed_navigation_leaks"`
	Unresolved              int `json:"unresolved"`
	DigestFailures          int `json:"digest_failures"`
	CorpusDrift             int `json:"corpus_drift"`
	ConceptDrift            int `json:"concept_drift"`
	HeadMismatches          int `json:"head_mismatches"`
	ProofFailures           int `json:"proof_failures"`
	RepositoryWrites        int `json:"repository_writes"`
	MutationAuthorities     int `json:"mutation_authorities"`
}
