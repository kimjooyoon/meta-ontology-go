package feedbackpredecessor

type Candidate struct {
	ArtifactID       int64  `json:"artifact_id"`
	RunID            int64  `json:"run_id"`
	RunAttempt       int    `json:"run_attempt"`
	ArtifactName     string `json:"artifact_name"`
	HeadSHA          string `json:"head_sha"`
	HeadBranch       string `json:"head_branch"`
	Workflow         string `json:"workflow"`
	Event            string `json:"event"`
	Conclusion       string `json:"conclusion"`
	Expired          bool   `json:"expired"`
	ReceiptDigest    string `json:"receipt_digest"`
	RepositoryWrites int    `json:"repository_writes"`
}

type Input struct {
	Repository        string      `json:"repository"`
	PredecessorSHA    string      `json:"predecessor_sha"`
	CanonicalBranch   string      `json:"canonical_branch"`
	CanonicalWorkflow string      `json:"canonical_workflow"`
	Candidates        []Candidate `json:"candidates"`
}

type Selection struct {
	ArtifactID    int64  `json:"artifact_id"`
	RunID         int64  `json:"run_id"`
	RunAttempt    int    `json:"run_attempt"`
	ReceiptDigest string `json:"receipt_digest"`
}

type Summary struct {
	ObservedCandidates     int `json:"observed_candidates"`
	ExactHeadCandidates    int `json:"exact_head_candidates"`
	CanonicalCandidates    int `json:"canonical_candidates"`
	SuccessfulCandidates   int `json:"successful_candidates"`
	AvailableCandidates    int `json:"available_candidates"`
	ReceiptBoundCandidates int `json:"receipt_bound_candidates"`
	AmbiguousCandidates    int `json:"ambiguous_candidates"`
	RepositoryWrites       int `json:"repository_writes"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	Target        int    `json:"target"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Activity      string `json:"activity"`
	Value         int    `json:"value"`
	Satisfied     bool   `json:"satisfied"`
}

type Report struct {
	Schema         string      `json:"schema"`
	Repository     string      `json:"repository"`
	PredecessorSHA string      `json:"predecessor_sha"`
	Decision       string      `json:"decision"`
	Reason         string      `json:"reason"`
	Selected       *Selection  `json:"selected,omitempty"`
	Summary        Summary     `json:"summary"`
	Indicators     []Indicator `json:"indicators"`
	ReportDigest   string      `json:"report_digest"`
}
