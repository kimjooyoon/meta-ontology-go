// Package proofchoicejudge is a separate consumer. It parses and lowers raw
// .gooo itself and never imports the producer package.
package proofchoicejudge

type slot struct {
	ID         string   `json:"id"`
	Observed   bool     `json:"observed"`
	Provenance []string `json:"provenance"`
}

type value struct {
	Schema           string   `json:"schema"`
	Kind             string   `json:"kind"`
	ID               string   `json:"id"`
	Statement        string   `json:"statement,omitempty"`
	PriorState       string   `json:"prior_state,omitempty"`
	Dependencies     []string `json:"dependencies,omitempty"`
	Observations     []string `json:"observations,omitempty"`
	AdmissibleRoutes []string `json:"admissible_routes,omitempty"`
	EvidenceKind     string   `json:"evidence_kind,omitempty"`
	Predicate        string   `json:"predicate,omitempty"`
	Value            string   `json:"value,omitempty"`
	Observed         bool     `json:"observed"`
	Provenance       []string `json:"provenance,omitempty"`
	Slots            []slot   `json:"slots,omitempty"`
}

type lowered struct {
	Values              []value
	Canonical           string
	SemanticDigest      string
	Reconstructed       int
	ReconstructionDenom int
}
