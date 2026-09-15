package causality

const (
	InputReportSchema = "gooo.language.operation-catalog/v3"
	ReceiptSchema     = "gooo.language.operation-claim-causality/v1"
	GraphSchema       = "gooo.language.operation-claim-dependency-contract/v1"
	ReceiptScope      = "CAUSAL_CLASSIFICATION_ONLY"

	ModeSuccess = "success"
	ModeUnknown = "unknown"

	ClaimTotal      = 9
	EdgeTotal       = 11
	TransitionTotal = 18
	IndicatorTotal  = 6
)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type inputReport struct {
	Schema                             string            `json:"schema"`
	SourceDigest                       string            `json:"source_digest"`
	CoreIRFingerprint                  string            `json:"core_ir_fingerprint"`
	OperationClaimTransitions          []inputTransition `json:"claim_transitions"`
	OperationClaimTransitionHead       string            `json:"operation_claim_transition_head"`
	OperationClaimTransitionHeadDigest string            `json:"claim_transition_head_digest"`
}

type inputTransition struct {
	Sequence                 int        `json:"sequence"`
	ClaimID                  string     `json:"claim_id"`
	DeclarationDigest        string     `json:"declaration_digest"`
	Event                    string     `json:"event"`
	Before                   string     `json:"before"`
	After                    string     `json:"after"`
	Coordinate               Coordinate `json:"coordinate"`
	EvidenceDigest           string     `json:"evidence_digest"`
	PreviousTransitionDigest string     `json:"previous_transition_digest"`
	TransitionDigest         string     `json:"transition_digest"`
}

type Subject struct {
	InputReportSchema      string   `json:"input_report_schema"`
	InputReportDigest      string   `json:"input_report_digest"`
	TransitionHeadDigest   string   `json:"transition_head_digest"`
	GraphDigest            string   `json:"graph_digest"`
	SourceDigest           string   `json:"source_digest"`
	SemanticIRDigest       string   `json:"semantic_ir_digest"`
	BindingStatus          string   `json:"binding_status"`
	MissingBindingEvidence []string `json:"missing_binding_evidence"`
}

type GraphNode struct {
	Ordinal int    `json:"ordinal"`
	Axis    string `json:"axis"`
	ClaimID string `json:"claim_id"`
}

type GraphEdge struct {
	EdgeID      string `json:"edge_id"`
	FromClaimID string `json:"from_claim_id"`
	ToClaimID   string `json:"to_claim_id"`
	Kind        string `json:"kind"`
}
