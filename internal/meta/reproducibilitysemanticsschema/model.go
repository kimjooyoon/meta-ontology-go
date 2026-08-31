package reproducibilitysemanticsschema

const (
	ReceiptSchema         = "gooo/reproducibility-semantics-receipt/v1"
	JudgmentSchema        = "gooo/reproducibility-semantics-judgment/v1"
	InterventionSchema    = "gooo/reproducibility-semantics-intervention/v1"
	ContractID            = "gooo/reproducibility-semantics/v1"
	ProducerID            = "gooo://meta/reproducibility-semantics/producer"
	ConsumerID            = "gooo://meta/reproducibility-semantics/judge"
	MetaOperation         = "separate-byte-and-meaning-claims"
	ProofByte             = "BYTE_COMPARISON"
	ProofMeaning          = "MEANING_ORACLE"
	ProofComposition      = "CLAIM_COMPOSITION"
	ProofSemantic         = "SEMANTIC_CAUSALITY"
	StatusOpen            = "OPEN"
	StatusDischarged      = "DISCHARGED"
	StatusRefuted         = "REFUTED"
	CaseCount             = 4
	InterventionCaseCount = 2
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

type Transition struct {
	From           string     `json:"from"`
	To             string     `json:"to"`
	Coordinate     Coordinate `json:"coordinate"`
	Stage          string     `json:"stage"`
	Step           string     `json:"step"`
	Reason         string     `json:"reason"`
	EvidenceDigest string     `json:"evidence_digest"`
}
