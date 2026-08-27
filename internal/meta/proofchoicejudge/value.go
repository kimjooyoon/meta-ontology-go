// Package proofchoicejudge is a separate consumer. It parses and lowers raw
// .gooo itself and never imports the producer package.
package proofchoicejudge

type value struct {
	Schema               string   `json:"schema"`
	Kind                 string   `json:"kind"`
	ID                   string   `json:"id"`
	Statement            string   `json:"statement,omitempty"`
	PriorState           string   `json:"prior_state,omitempty"`
	Subject              string   `json:"subject,omitempty"`
	EvidenceCapabilities []string `json:"evidence_capabilities,omitempty"`
	Members              []string `json:"members,omitempty"`
}

type lowered struct {
	Values              []value
	Bindings            map[string]string
	Canonical           string
	SemanticDigest      string
	Reconstructed       int
	ReconstructionDenom int
}
