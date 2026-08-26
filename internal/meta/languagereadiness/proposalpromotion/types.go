package proposalpromotion

const (
	Schema             = "gooo/autonomous-change-proposal-promotion/v2"
	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
	ReasonReady        = "MERGED_CHANGE_PROPOSAL_PROMOTION_READY"
	ReasonUnresolved   = "MERGED_CHANGE_PROPOSAL_PROMOTION_UNRESOLVED"
	totalCoordinates   = 8
)

type Receipt struct {
	Schema                       string       `json:"schema"`
	Repository                   string       `json:"repository"`
	CurrentHeadSHA               string       `json:"current_head_sha"`
	EvidenceHeadSHA              string       `json:"evidence_head_sha"`
	Decision                     string       `json:"decision"`
	Reason                       string       `json:"reason"`
	MetaOperation                string       `json:"meta_operation"`
	Source                       Source       `json:"source"`
	Summary                      Summary      `json:"summary"`
	Coordinates                  []Coordinate `json:"coordinates"`
	Indicators                   []Indicator  `json:"indicators"`
	Proofs                       []Proof      `json:"proofs"`
	RepositoryWrites             int          `json:"repository_writes"`
	RepositoryMutationAuthorized bool         `json:"repository_mutation_authorized"`
	ReportDigest                 string       `json:"report_digest"`
}

type Summary struct {
	Satisfied           int `json:"satisfied"`
	Total               int `json:"total"`
	NotSatisfied        int `json:"not_satisfied"`
	Unresolved          int `json:"unresolved"`
	ReadinessBPS        int `json:"readiness_bps"`
	ValidPredecessors   int `json:"valid_predecessors"`
	AmbiguousCandidates int `json:"ambiguous_candidates"`
	RepositoryWrites    int `json:"repository_writes"`
}

type Coordinate struct {
	ID             string `json:"id"`
	ProofChoice    string `json:"proof_choice"`
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
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
	Passed         bool   `json:"passed"`
	EvidenceDigest string `json:"evidence_digest"`
}
