package languagepackageruntime

type Summary struct {
	Total                 int `json:"total"`
	Executed              int `json:"executed"`
	Satisfied             int `json:"satisfied"`
	NotSatisfied          int `json:"not_satisfied"`
	Unresolved            int `json:"unresolved"`
	PositivePaths         int `json:"positive_paths"`
	GuardrailRejections   int `json:"guardrail_rejections"`
	Packages              int `json:"packages"`
	Sources               int `json:"sources"`
	Imports               int `json:"imports"`
	Initializations       int `json:"initializations"`
	EntryBindings         int `json:"entry_bindings"`
	SemanticBindings      int `json:"semantic_bindings"`
	CanonicalReplays      int `json:"canonical_replays"`
	OrderInvariantReplays int `json:"order_invariant_replays"`
	UnknownObservations   int `json:"unknown_observations"`
	InvalidAcceptances    int `json:"invalid_acceptances"`
	GraphAcceptances      int `json:"graph_acceptances"`
	SourceAcceptances     int `json:"source_acceptances"`
	EntryAcceptances      int `json:"entry_acceptances"`
	EffectfulOperations   int `json:"effectful_operations"`
	RepositoryWrites      int `json:"repository_writes"`
	MutationAuthorities   int `json:"mutation_authorities"`
	MetricBindings        int `json:"metric_bindings"`
}
