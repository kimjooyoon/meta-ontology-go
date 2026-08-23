package promotioncontinuity

type Summary struct {
	Satisfied                 int `json:"satisfied"`
	Total                     int `json:"total"`
	NotSatisfied              int `json:"not_satisfied"`
	Unresolved                int `json:"unresolved"`
	ReadinessBPS              int `json:"readiness_bps"`
	AuthorizedGuardReceipts   int `json:"authorized_guard_receipts"`
	AuthorizedRecoveryRoutes  int `json:"authorized_recovery_routes"`
	RepositoryWrites          int `json:"repository_writes"`
}

type Coordinate struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}

type Indicator struct {
	MetricID     string `json:"metric_id"`
	Class        string `json:"class"`
	ProofChoice  string `json:"proof_choice"`
	Producer     string `json:"producer"`
	Consumer     string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Resolution   string `json:"resolution"`
	Value        int    `json:"value"`
	Target       int    `json:"target"`
	Satisfied    bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Report struct {
	Schema                       string       `json:"schema"`
	Decision                     string       `json:"decision"`
	Reason                       string       `json:"reason"`
	Resolution                   string       `json:"resolution"`
	Producer                     string       `json:"producer"`
	Consumer                     string       `json:"consumer"`
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
