package claimledger

type Evidence struct {
	ID                  string `json:"id"`
	ClaimID             string `json:"claim_id"`
	Status              string `json:"status"`
	SourcePath          string `json:"source_path"`
	SourceDigest        string `json:"source_digest"`
	ObservedValueDigest string `json:"observed_value_digest,omitempty"`
	ExpectedValueDigest string `json:"expected_value_digest,omitempty"`
}

type Claim struct {
	ID           string     `json:"id"`
	Kind         string     `json:"kind"`
	Modality     string     `json:"modality"`
	Subject      string     `json:"subject"`
	Predicate    string     `json:"predicate"`
	Scope        string     `json:"scope"`
	ProofRoute   string     `json:"proof_route"`
	Coordinate   Coordinate `json:"coordinate"`
	Status       string     `json:"status"`
	Truth        string     `json:"truth"`
	Reason       string     `json:"reason"`
	EvidenceRefs []string   `json:"evidence_refs"`
}

type Event struct {
	Sequence   int        `json:"sequence"`
	Type       string     `json:"type"`
	ClaimID    string     `json:"claim_id"`
	EvidenceID string     `json:"evidence_id,omitempty"`
	Status     string     `json:"status"`
	Coordinate Coordinate `json:"coordinate"`
	Reason     string     `json:"reason"`
}
