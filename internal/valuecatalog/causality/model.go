package causality

const (
	InputReportSchema = "gooo.language.operation-catalog/v3"
	ReceiptSchema     = "gooo.language.operation-claim-causality/v1"
	GraphSchema       = "gooo.language.operation-claim-dependency-contract/v1"
	ReceiptScope      = "CAUSAL_CLASSIFICATION_ONLY"

	ModeSuccess = "success"
	ModeUnknown = "unknown"

	ClaimTotal     = 9
	EdgeTotal      = 11
	TransitionTotal = 18
	IndicatorTotal = 6
)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type inputReport struct {
	Schema                             string            `json:"schema"`
	OperationClaimTransitions          []inputTransition `json:"operation_claim_transitions"`
	OperationClaimTransitionHead       string            `json:"operation_claim_transition_head"`
	OperationClaimTransitionHeadDigest string            `json:"operation_claim_transition_head_digest"`
}

type inputTransition struct {
	Sequence                     int        `json:"sequence"`
	ClaimID                      string     `json:"claim_id"`
	DeclarationDigest            string     `json:"declaration_digest"`
	Event                        string     `json:"event"`
	Before                       string     `json:"before"`
	After                        string     `json:"after"`
	Coordinate                   Coordinate `json:"coordinate"`
	EvidenceDigest               string     `json:"evidence_digest"`
	PreviousTransitionDigest     string     `json:"previous_transition_digest"`
	TransitionDigest             string     `json:"transition_digest"`
}

type Subject struct {
	InputReportSchema       string   `json:"input_report_schema"`
	InputReportDigest       string   `json:"input_report_digest"`
	TransitionHeadDigest    string   `json:"transition_head_digest"`
	GraphDigest             string   `json:"graph_digest"`
	BindingStatus           string   `json:"binding_status"`
	MissingBindingEvidence  []string `json:"missing_binding_evidence"`
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

type GraphContract struct {
	Schema                       string      `json:"schema"`
	Authority                    string      `json:"authority"`
	Completeness                 string      `json:"completeness"`
	SemanticCorrectnessClaimed   bool        `json:"semantic_correctness_claimed"`
	NodeTotal                    int         `json:"node_total"`
	EdgeTotal                    int         `json:"edge_total"`
	Nodes                        []GraphNode `json:"nodes"`
	Edges                        []GraphEdge `json:"edges"`
	Digest                       string      `json:"digest"`
}

type Resolution struct {
	ClaimID               string     `json:"claim_id"`
	Axis                  string     `json:"axis"`
	State                 string     `json:"state"`
	Kind                  string     `json:"kind"`
	ObservedEvent         string     `json:"observed_event"`
	Coordinate            Coordinate `json:"coordinate"`
	MissingEvidenceIDs    []string   `json:"missing_evidence_ids"`
	BlockedByClaimIDs     []string   `json:"blocked_by_claim_ids"`
	BlockedByEdgeIDs      []string   `json:"blocked_by_edge_ids"`
	CausePath             []string   `json:"cause_path"`
	CauseTransitionDigest string     `json:"cause_transition_digest,omitempty"`
	CauseCoordinate       *Coordinate `json:"cause_coordinate,omitempty"`
}

type Metrics struct {
	ContractClaimTotal          int `json:"contract_claim_total"`
	ContractEdgeTotal           int `json:"contract_edge_total"`
	ClassifiedClaimTotal        int `json:"classified_claim_total"`
	DischargedClaimTotal        int `json:"discharged_claim_total"`
	UnknownClaimTotal           int `json:"unknown_claim_total"`
	DirectMissingClaimTotal     int `json:"direct_missing_claim_total"`
	DependencyBlockedClaimTotal int `json:"dependency_blocked_claim_total"`
	ObservedBlockingEdgeTotal   int `json:"observed_blocking_edge_total"`
	MaximumCausePathDepth       int `json:"maximum_cause_path_depth"`
	ClassificationBasisPoints  int `json:"classification_basis_points"`
	DischargeBasisPoints       int `json:"discharge_basis_points"`
}

type Indicator struct {
	IndicatorID string `json:"indicator_id"`
	Class       string `json:"class"`
	Trilemma    string `json:"trilemma"`
	Value       int    `json:"value"`
	Target      int    `json:"target"`
	Comparator  string `json:"comparator"`
	Satisfied   bool   `json:"satisfied"`
}

type Decision struct {
	Value                       string `json:"value"`
	Resolution                  string `json:"resolution"`
	Reason                      string `json:"reason"`
	SemanticPromotionAuthorized bool   `json:"semantic_promotion_authorized"`
}

type Receipt struct {
	Schema        string        `json:"schema"`
	Scope         string        `json:"scope"`
	Subject       Subject       `json:"subject"`
	Graph         GraphContract `json:"graph"`
	Metrics       Metrics       `json:"metrics"`
	Indicators    []Indicator   `json:"indicators"`
	Resolutions   []Resolution  `json:"resolutions"`
	Decision      Decision      `json:"decision"`
	ReceiptDigest string        `json:"receipt_digest"`
}
