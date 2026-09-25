package languagesourcebindingpromotion

const (
	ReportSchema        = "gooo/language-source-binding-promotion/v1"
	ContractSchema      = "gooo/language-source-binding-promotion-contract/v1"
	IndependenceSchema  = "gooo/language-source-binding-promotion-independence/v1"
	Scope               = "SOURCE_EXECUTION_RECEIPT_BINDING_ONLY"
	DecisionPass        = "PASS"
	DecisionClosed      = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"
)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type IndependenceEvidence struct {
	Schema               string `json:"schema"`
	ProducerDependencies int    `json:"producer_dependencies"`
}

type Input struct {
	Contract             Contract
	HeadSHA              string
	PolicySource         []byte
	PolicyArtifact       []byte
	PolicyReplayArtifact []byte
	Producer             []byte
	Receipt              []byte
	Oracle               []byte
	UnknownProducer      []byte
	UnknownOracle        []byte
	MismatchedOracle     []byte
	Independence         IndependenceEvidence
}

type ClaimDefinition struct {
	ID            string `json:"id"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
}

type EdgeDefinition struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type CaseDefinition struct {
	ID                      string `json:"id"`
	ExpectedDecision        string `json:"expected_decision"`
	ExpectedResolution      string `json:"expected_resolution"`
	ExpectedReason          string `json:"expected_reason"`
	ExpectedPromotionStatus string `json:"expected_promotion_status"`
}

type Contract struct {
	Schema string            `json:"schema"`
	Scope  string            `json:"scope"`
	Claims []ClaimDefinition `json:"claims"`
	Edges  []EdgeDefinition  `json:"edges"`
	Cases  []CaseDefinition  `json:"cases"`
}
