package causality

type GraphContract struct {
	Schema                     string      `json:"schema"`
	Authority                  string      `json:"authority"`
	Completeness               string      `json:"completeness"`
	SemanticCorrectnessClaimed bool        `json:"semantic_correctness_claimed"`
	NodeTotal                  int         `json:"node_total"`
	EdgeTotal                  int         `json:"edge_total"`
	Nodes                      []GraphNode `json:"nodes"`
	Edges                      []GraphEdge `json:"edges"`
	Digest                     string      `json:"digest"`
}

type Resolution struct {
	ClaimID               string      `json:"claim_id"`
	Axis                  string      `json:"axis"`
	State                 string      `json:"state"`
	Kind                  string      `json:"kind"`
	ObservedEvent         string      `json:"observed_event"`
	Coordinate            Coordinate  `json:"coordinate"`
	MissingEvidenceIDs    []string    `json:"missing_evidence_ids"`
	BlockedByClaimIDs     []string    `json:"blocked_by_claim_ids"`
	BlockedByEdgeIDs      []string    `json:"blocked_by_edge_ids"`
	CausePath             []string    `json:"cause_path"`
	CauseTransitionDigest string      `json:"cause_transition_digest,omitempty"`
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
	ClassificationBasisPoints   int `json:"classification_basis_points"`
	DischargeBasisPoints        int `json:"discharge_basis_points"`
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
