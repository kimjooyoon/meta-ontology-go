package verticalsliceclosureshadow

type BoundaryResult struct {
	ID                 string `json:"id"`
	Schema             string `json:"schema"`
	MetaOperation      string `json:"meta_operation"`
	Status             string `json:"status"`
	Resolution         string `json:"resolution"`
	Reason             string `json:"reason"`
	Value              int    `json:"value"`
	Target             int    `json:"target"`
	LinksSatisfied     int    `json:"links_satisfied"`
	LinksTotal         int    `json:"links_total"`
	HeadSHA            string `json:"head_sha,omitempty"`
	EvidenceDigest     string `json:"evidence_digest,omitempty"`
	ReportDigest       string `json:"report_digest,omitempty"`
	EvidenceAvailable  bool   `json:"evidence_available"`
	UnknownTopDecision bool   `json:"unknown_top_decision"`
	KnownFailure       bool   `json:"known_failure"`
	RepositoryWrites   int    `json:"repository_writes"`
}

type Summary struct {
	DenominatorTotal         int `json:"denominator_total"`
	BeforeOperating          int `json:"before_operating"`
	ProjectedOperating       int `json:"projected_operating"`
	BeforeCoverageBPS        int `json:"before_coverage_bps"`
	ProjectedCoverageBPS     int `json:"projected_coverage_bps"`
	BoundariesTotal          int `json:"boundaries_total"`
	BoundariesSatisfied      int `json:"boundaries_satisfied"`
	UnknownBoundaries        int `json:"unknown_boundaries"`
	BlockedBoundaries        int `json:"blocked_boundaries"`
	BoundaryCoverageBPS      int `json:"boundary_coverage_bps"`
	LinksTotal               int `json:"links_total"`
	LinksSatisfied           int `json:"links_satisfied"`
	LinkCoverageBPS          int `json:"link_coverage_bps"`
	EvidenceAvailable        int `json:"evidence_available"`
	UnknownTopDecisions      int `json:"unknown_top_decisions"`
	KnownFailures            int `json:"known_failures"`
	ObservedRepositoryWrites int `json:"observed_repository_writes"`
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

type Report struct {
	Schema              string           `json:"schema"`
	MetricID            string           `json:"metric_id"`
	MetaOperation       string           `json:"meta_operation"`
	Decision            string           `json:"decision"`
	Reason              string           `json:"reason"`
	Resolution          string           `json:"resolution"`
	EnforcementEffect   string           `json:"enforcement_effect"`
	HeadSHA             string           `json:"head_sha"`
	AssuranceSubjectSHA string           `json:"assurance_subject_sha,omitempty"`
	AssuranceDigest     string           `json:"assurance_digest"`
	DenominatorDigest   string           `json:"denominator_digest"`
	Summary             Summary          `json:"summary"`
	Boundaries          []BoundaryResult `json:"boundaries"`
	Indicators          []Indicator      `json:"indicators"`
	RepositoryWrites    int              `json:"repository_writes"`
	PromotionApplied    int              `json:"promotion_applied"`
	ReportDigest        string           `json:"report_digest"`
}
