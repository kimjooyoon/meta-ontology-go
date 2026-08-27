package reproducibilitysemantics

const (
	ReceiptSchema    = "gooo/reproducibility-semantics-receipt/v1"
	JudgmentSchema   = "gooo/reproducibility-semantics-judgment/v1"
	ContractID       = "gooo/reproducibility-semantics/v1"
	ProducerID       = "gooo://meta/reproducibility-semantics/producer"
	ConsumerID       = "gooo://meta/reproducibility-semantics/judge"
	MetaOperation    = "separate-byte-and-meaning-claims"
	ProofByte        = "BYTE_COMPARISON"
	ProofMeaning     = "MEANING_ORACLE"
	ProofComposition = "CLAIM_COMPOSITION"
	StatusOpen       = "OPEN"
	StatusDischarged = "DISCHARGED"
	StatusRefuted    = "REFUTED"
	CaseCount        = 4
)

type Coordinate struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
	BasisPoints int `json:"basis_points"`
}

type Evidence struct {
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Reference     string `json:"reference"`
	Candidate     string `json:"candidate"`
	Status        string `json:"status"`
}

type MeaningEvidence struct {
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Expected      string `json:"expected"`
	Observed      string `json:"observed"`
	Status        string `json:"status"`
}

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
