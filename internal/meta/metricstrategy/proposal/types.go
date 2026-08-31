package proposal

const (
	Schema         = "gooo/autonomous-change-proposal-contract/v1"
	RegistrySchema = "gooo/autonomous-change-proposal-coordinate-registry/v1"
)

type CoordinateSpec struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
}

type Coordinate struct {
	CoordinateSpec
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Summary struct {
	Satisfied        int `json:"satisfied"`
	Total            int `json:"total"`
	NotSatisfied     int `json:"not_satisfied"`
	Unresolved       int `json:"unresolved"`
	ReadinessBPS     int `json:"readiness_bps"`
	RatioNumerator   int `json:"ratio_numerator"`
	RatioDenominator int `json:"ratio_denominator"`
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
	Unit          string `json:"unit"`
	Satisfied     bool   `json:"satisfied"`
}

type Report struct {
	Schema              string       `json:"schema"`
	RegistrySchema      string       `json:"registry_schema"`
	RegistryDigest      string       `json:"registry_digest"`
	Repository          string       `json:"repository"`
	SubjectSHA          string       `json:"subject_sha"`
	Decision            string       `json:"decision"`
	Reason              string       `json:"reason"`
	StrategyDecision    string       `json:"strategy_decision"`
	ProposalDecision    string       `json:"proposal_decision"`
	SelectedActions     int          `json:"selected_actions"`
	Summary             Summary      `json:"summary"`
	Coordinates         []Coordinate `json:"coordinates"`
	Indicators          []Indicator  `json:"indicators"`
	Proofs              []Proof      `json:"proofs"`
	RepositoryWrites    int          `json:"repository_writes"`
	PromotionAuthorized bool         `json:"promotion_authorized"`
	ReportDigest        string       `json:"report_digest"`
}
