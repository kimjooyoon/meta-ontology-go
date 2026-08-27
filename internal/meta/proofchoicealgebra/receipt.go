package proofchoicealgebra

type Item struct {
	Kind           string   `json:"kind"`
	ID             string   `json:"id"`
	Statement      string   `json:"statement,omitempty"`
	PriorState     string   `json:"prior_state,omitempty"`
	Choice         Route    `json:"choice,omitempty"`
	Resolution     string   `json:"resolution"`
	Observations   []string `json:"observations"`
	Numerator      int      `json:"numerator,omitempty"`
	Denominator    int      `json:"denominator,omitempty"`
	EvidenceDigest string   `json:"evidence_digest"`
	Provenance     []string `json:"provenance"`
}

type Transition struct {
	Sequence       int      `json:"sequence"`
	ClaimID        string   `json:"claim_id"`
	From           string   `json:"from"`
	To             string   `json:"to"`
	Choice         Route    `json:"choice,omitempty"`
	Resolution     string   `json:"resolution"`
	Stage          string   `json:"stage"`
	Step           string   `json:"step"`
	Reason         string   `json:"reason"`
	EvidenceDigest string   `json:"evidence_digest"`
	Provenance     []string `json:"provenance"`
	Persistent     bool     `json:"persistent"`
}
