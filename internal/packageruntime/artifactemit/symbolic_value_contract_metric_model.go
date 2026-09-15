package artifactemit

type SymbolicValueContractIndicator struct {
	ID            string   `json:"id"`
	Class         string   `json:"class"`
	ProofChoice   string   `json:"proof_choice"`
	MetaOperation string   `json:"meta_operation"`
	Observed      int      `json:"observed"`
	Expected      int      `json:"expected"`
	Satisfied     bool     `json:"satisfied"`
	Audiences     []string `json:"audiences"`
}

type SymbolicValueContractView struct {
	Audience    string `json:"audience"`
	Resolution  string `json:"resolution"`
	Satisfied   int    `json:"satisfied"`
	Total       int    `json:"total"`
	BasisPoints int    `json:"basis_points"`
}

type SymbolicValueContractProof struct {
	ProofChoice string `json:"proof_choice"`
	Satisfied   int    `json:"satisfied"`
	Total       int    `json:"total"`
}

type SymbolicValueContractEffects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}
