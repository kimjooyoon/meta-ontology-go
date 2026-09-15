package guardedpromotion

type Coordinate struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	Satisfied      bool   `json:"satisfied"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Summary struct {
	Satisfied                    int  `json:"satisfied"`
	Total                        int  `json:"total"`
	NotSatisfied                 int  `json:"not_satisfied"`
	Unresolved                   int  `json:"unresolved"`
	ReadinessBPS                 int  `json:"readiness_bps"`
	ValidPredecessors            int  `json:"valid_predecessors"`
	AmbiguousCandidates          int  `json:"ambiguous_candidates"`
	RepositoryWrites             int  `json:"repository_writes"`
	ReadinessPromotionAuthorized bool `json:"readiness_promotion_authorized"`
	RepositoryMutationAuthorized bool `json:"repository_mutation_authorized"`
}

type Report struct {
	Schema       string       `json:"schema"`
	Decision     string       `json:"decision"`
	Reason       string       `json:"reason"`
	Resolution   string       `json:"resolution"`
	Source       Source       `json:"source"`
	Summary      Summary      `json:"summary"`
	Coordinates  []Coordinate `json:"coordinates"`
	Indicators   []Indicator  `json:"indicators"`
	Proofs       []Proof      `json:"proofs"`
	ReportDigest string       `json:"report_digest"`
}
