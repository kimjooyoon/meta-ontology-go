package selfimprovementattestation

type Coordinate struct {
	Stage string `json:"stage"`
	Step  string `json:"step"`
}

type Contract struct {
	ContractID      string `json:"contract_id"`
	Path            string `json:"path"`
	Package         string `json:"package"`
	Namespace       string `json:"namespace"`
	EntityCount     int    `json:"entity_count"`
	ActivityCount   int    `json:"activity_count"`
	ObligationTotal int    `json:"obligation_total"`
	SourceLines     int    `json:"source_lines"`
	SourceDigest    string `json:"source_digest"`
	CanonicalDigest string `json:"canonical_digest"`
}

type Metrics struct {
	FixedObligationTotal int `json:"fixed_obligation_total"`
	VerifiedTotal        int `json:"verified_total"`
	UnknownTotal         int `json:"unknown_total"`
	FalseTotal           int `json:"false_total"`
	OpenTotal            int `json:"open_total"`
	CoverageBasisPoints  int `json:"coverage_basis_points"`
	FalsePromotionCount  int `json:"false_promotion_count"`
}

type Obligation struct {
	ID             string     `json:"id"`
	ProofRoute     string     `json:"proof_route"`
	Coordinate     Coordinate `json:"coordinate"`
	Status         string     `json:"status"`
	Reason         string     `json:"reason"`
	EvidenceDigest string     `json:"evidence_digest,omitempty"`
}
