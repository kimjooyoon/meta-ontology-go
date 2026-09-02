package claimledger

type Metrics struct {
	FixedClaimTotal      int              `json:"fixed_claim_total"`
	InScopeClaimTotal    int              `json:"in_scope_claim_total"`
	DischargedTotal      int              `json:"discharged_total"`
	UnknownTotal         int              `json:"unknown_total"`
	RefutedTotal         int              `json:"refuted_total"`
	ExcludedTotal        int              `json:"excluded_total"`
	OpenClaimTotal       int              `json:"open_claim_total"`
	DischargeBasisPoints int              `json:"discharge_basis_points"`
	FalsePromotionCount  int              `json:"false_promotion_count"`
	ProofRoutes          ProofRouteCounts `json:"proof_routes"`
}

type Verdict struct {
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
}

type Conformance struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type Report struct {
	Schema            string        `json:"schema"`
	Subject           string        `json:"subject"`
	Metric            string        `json:"metric"`
	ContractDigest    string        `json:"contract_digest"`
	ObservationDigest string        `json:"observation_digest"`
	RuntimeDigest     string        `json:"runtime_evidence_digest"`
	Inputs            []InputRecord `json:"inputs"`
	ClaimSet          Verdict       `json:"claim_set"`
	Conformance       Conformance   `json:"conformance"`
	Metrics           Metrics       `json:"metrics"`
	OpenClaimIDs      []string      `json:"open_claim_ids"`
	Claims            []Claim       `json:"claims"`
	Evidence          []Evidence    `json:"evidence"`
	Events            []Event       `json:"events"`
}
