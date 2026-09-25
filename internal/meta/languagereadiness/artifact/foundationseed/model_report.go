package foundationseed

type Summary struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Authority struct {
	RepositoryMutationAuthorized bool `json:"repository_mutation_authorized"`
	ReadinessDeltaAuthorized     bool `json:"readiness_delta_authorized"`
	PromotionAuthorized          bool `json:"promotion_authorized"`
	AutomaticAdoptionAuthorized  bool `json:"automatic_adoption_authorized"`
}

type Report struct {
	Schema      string      `json:"schema"`
	Conformance string      `json:"conformance"`
	Decision    string      `json:"decision"`
	Reason      string      `json:"reason"`
	Resolution  string      `json:"resolution"`
	Source      Source      `json:"source"`
	Summary     Summary     `json:"summary"`
	Indicators  []Indicator `json:"indicators"`
	Views       []View      `json:"views"`
	Proofs      []Proof     `json:"proofs"`
	NonClaims   []string    `json:"non_claims"`
	Authority   Authority   `json:"authority"`
	Digest      string      `json:"digest"`
}
