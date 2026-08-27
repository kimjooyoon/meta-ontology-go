package proofchoicealgebra

// Value is the source-side proposition and evidence capability contract. It
// contains no observed result, route choice, or metric slot.
type Value struct {
	Schema               string   `json:"schema"`
	Kind                 string   `json:"kind"`
	ID                   string   `json:"id"`
	Statement            string   `json:"statement,omitempty"`
	PriorState           string   `json:"prior_state,omitempty"`
	Subject              string   `json:"subject,omitempty"`
	EvidenceCapabilities []Route  `json:"evidence_capabilities,omitempty"`
	Members              []string `json:"members,omitempty"`
}

type lowered struct {
	Values              []Value
	Bindings            map[string]string
	Canonical           string
	SemanticDigest      string
	Reconstructed       int
	ReconstructionDenom int
}

type routeResult struct {
	Route            Route
	Resolution       string
	ObservationState string
	Reason           string
	Observations     []string
	EvidenceDigest   string
	Provenance       []string
	FailClosed       bool
}
