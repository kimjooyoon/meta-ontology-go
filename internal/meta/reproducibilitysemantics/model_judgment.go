package reproducibilitysemantics

type JudgmentCase struct {
	ID            string `json:"id"`
	ByteStatus    string `json:"byte_status"`
	MeaningStatus string `json:"meaning_status"`
	Status        string `json:"status"`
	Reason        string `json:"reason"`
}

type Judgment struct {
	Schema         string         `json:"schema"`
	Version        int            `json:"version"`
	ContractID     string         `json:"contract_id"`
	SourcePath     string         `json:"source_path"`
	SourceDigest   string         `json:"source_digest"`
	SemanticDigest string         `json:"semantic_digest"`
	HeadSHA        string         `json:"head_sha"`
	ReceiptDigest  string         `json:"receipt_digest"`
	Producer       string         `json:"producer"`
	Consumer       string         `json:"consumer"`
	MetaOperation  string         `json:"meta_operation"`
	ProofChoice    string         `json:"proof_choice"`
	Stage          string         `json:"stage"`
	Step           string         `json:"step"`
	Reason         string         `json:"reason"`
	Decision       string         `json:"decision"`
	Cases          []JudgmentCase `json:"cases"`
	Summary        Summary        `json:"summary"`
	Proofs         []Proof        `json:"proofs"`
	Authority      Authority      `json:"authority"`
	JudgmentDigest string         `json:"judgment_digest"`
}
