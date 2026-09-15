package languagepackageruntime

type CaseResult struct {
	ID                    string `json:"id"`
	Kind                  string `json:"kind"`
	Expected              string `json:"expected"`
	Observed              string `json:"observed"`
	Reason                string `json:"reason"`
	RuntimeDigest         string `json:"runtime_digest"`
	Satisfied             bool   `json:"satisfied"`
	Packages              int    `json:"packages"`
	Sources               int    `json:"sources"`
	Imports               int    `json:"imports"`
	Initializations       int    `json:"initializations"`
	EntryBindings         int    `json:"entry_bindings"`
	SemanticBindings      int    `json:"semantic_bindings"`
	CanonicalReplays      int    `json:"canonical_replays"`
	OrderInvariantReplays int    `json:"order_invariant_replays"`
	Effects               int    `json:"effects"`
	RepositoryWrites      int    `json:"repository_writes"`
}
