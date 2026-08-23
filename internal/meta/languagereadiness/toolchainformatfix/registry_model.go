package toolchainformatfix

type Definition struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Operation     string `json:"operation"`
	ExpectedExit  int    `json:"expected_exit"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
}

type Registry struct {
	Schema  string       `json:"schema"`
	Version string       `json:"version"`
	Cases   []Definition `json:"cases"`
}
