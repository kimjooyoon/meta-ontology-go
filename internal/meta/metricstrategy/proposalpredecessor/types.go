package proposalpredecessor

const Schema = "gooo/autonomous-change-proposal-predecessor-selection/v2"

type Selected struct {
	RunID                  int64  `json:"run_id"`
	RunAttempt             int    `json:"run_attempt"`
	HeadSHA                string `json:"head_sha"`
	Event                  string `json:"event"`
	Status                 string `json:"status"`
	Conclusion             string `json:"conclusion"`
	WorkflowName           string `json:"workflow_name"`
	SynthesisJobID         int64  `json:"synthesis_job_id"`
	SynthesisJobName       string `json:"synthesis_job_name"`
	SynthesisJobStatus     string `json:"synthesis_job_status"`
	SynthesisJobConclusion string `json:"synthesis_job_conclusion"`
	ArtifactID             int64  `json:"artifact_id"`
	ArtifactName           string `json:"artifact_name"`
	ProposalFileSHA256     string `json:"proposal_file_sha256"`
	ProposalReportDigest   string `json:"proposal_report_digest"`
	ContractSatisfied      int    `json:"contract_satisfied"`
	ContractTotal          int    `json:"contract_total"`
	ContractBPS            int    `json:"contract_bps"`
	ContractUnresolved     int    `json:"contract_unresolved"`
	RepositoryWrites       int    `json:"repository_writes"`
	PromotionAuthorized    bool   `json:"promotion_authorized"`
}

type Candidate struct {
	Selected
	ProposalPayload []byte `json:"-"`
}

type Collection struct {
	ObservedRuns      int
	ExactRuns         int
	ObservedArtifacts int
	ExactArtifacts    int
	ObservedJobs      int
	ExactJobs         int
	Unresolved        int
	Candidates        []Candidate
}

type Report struct {
	Schema            string      `json:"schema"`
	Repository        string      `json:"repository"`
	CurrentSubjectSHA string      `json:"current_subject_sha"`
	PredecessorSHA    string      `json:"predecessor_sha"`
	Decision          string      `json:"decision"`
	Reason            string      `json:"reason"`
	Selected          *Selected   `json:"selected,omitempty"`
	Summary           Summary     `json:"summary"`
	Indicators        []Indicator `json:"indicators"`
	Proofs            []Proof     `json:"proofs"`
	ReportDigest      string      `json:"report_digest"`
}

func (report Report) Ready() bool {
	return report.Decision == "SELECTED" && report.Reason == "PROPOSAL_PREDECESSOR_SELECTED"
}
