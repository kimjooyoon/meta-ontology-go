package reproducibilitysemantics

type Case struct {
	ID      string          `json:"id"`
	Byte    Evidence        `json:"byte"`
	Meaning MeaningEvidence `json:"meaning"`
	Status  string          `json:"status"`
	Stage   string          `json:"stage"`
	Step    string          `json:"step"`
	Reason  string          `json:"reason"`
}

type Summary struct {
	CaseMatrix      Coordinate `json:"case_matrix"`
	ByteClaim       Coordinate `json:"byte_claim"`
	MeaningClaim    Coordinate `json:"meaning_claim"`
	JointClaim      Coordinate `json:"joint_claim"`
	Counterexamples Coordinate `json:"counterexamples"`
	OpenCases       Coordinate `json:"open_cases"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
	Status         string `json:"status"`
}

type Authority struct {
	RepositoryWrites    int  `json:"repository_writes"`
	MutationAuthorized  bool `json:"mutation_authorized"`
	PromotionAuthorized bool `json:"promotion_authorized"`
}

type Receipt struct {
	Schema        string    `json:"schema"`
	Version       int       `json:"version"`
	ContractID    string    `json:"contract_id"`
	SourcePath    string    `json:"source_path"`
	SourceDigest  string    `json:"source_digest"`
	HeadSHA       string    `json:"head_sha"`
	Producer      string    `json:"producer"`
	Consumer      string    `json:"consumer"`
	MetaOperation string    `json:"meta_operation"`
	ProofChoice   string    `json:"proof_choice"`
	Stage         string    `json:"stage"`
	Step          string    `json:"step"`
	Reason        string    `json:"reason"`
	Cases         []Case    `json:"cases"`
	Summary       Summary   `json:"summary"`
	Proofs        []Proof   `json:"proofs"`
	Authority     Authority `json:"authority"`
	ReceiptDigest string    `json:"receipt_digest"`
}
