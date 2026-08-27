package proofchoicealgebra

// Item is a claim or metric carrying one explicit proof-choice meta value.
// The fields describe evidence routing, not a theorem-prover type system.
type Item struct {
	Kind          Kind   `json:"kind"`
	ID            string `json:"id"`
	Statement     string `json:"statement"`
	Choice        Choice `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Numerator     int    `json:"numerator,omitempty"`
	Denominator   int    `json:"denominator,omitempty"`
	Line          int    `json:"line"`
}

// Transition records a persistent claim moving between lifecycle states. Its
// choice is immutable across the transition and is checked against ClaimID.
type Transition struct {
	ClaimID       string `json:"claim_id"`
	From          string `json:"from"`
	To            string `json:"to"`
	Choice        Choice `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Persistent    bool   `json:"persistent"`
	Line          int    `json:"line"`
}

type Bundle struct {
	Items       []Item
	Transitions []Transition
}
