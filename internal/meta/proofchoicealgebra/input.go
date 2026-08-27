package proofchoicealgebra

type Slot struct {
	ID         string   `json:"id"`
	Observed   bool     `json:"observed"`
	Provenance []string `json:"provenance"`
}

// Value is the only semantic carrier accepted from a computes value program.
// Choice, numerator, and denominator are deliberately absent from the input.
type Value struct {
	Schema           string   `json:"schema"`
	Kind             string   `json:"kind"`
	ID               string   `json:"id"`
	Statement        string   `json:"statement,omitempty"`
	PriorState       string   `json:"prior_state,omitempty"`
	Dependencies     []string `json:"dependencies,omitempty"`
	Observations     []string `json:"observations,omitempty"`
	AdmissibleRoutes []Route  `json:"admissible_routes,omitempty"`
	EvidenceKind     string   `json:"evidence_kind,omitempty"`
	Predicate        string   `json:"predicate,omitempty"`
	Value            string   `json:"value,omitempty"`
	Observed         bool     `json:"observed"`
	Provenance       []string `json:"provenance,omitempty"`
	Slots            []Slot   `json:"slots,omitempty"`
}

type lowered struct {
	Values              []Value
	Canonical           string
	SemanticDigest      string
	Reconstructed       int
	ReconstructionDenom int
}

type routeResult struct {
	Route        Route
	Resolution   string
	Reason       string
	Observations []string
}
