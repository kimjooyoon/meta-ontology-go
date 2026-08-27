package proposalpredecessor

const (
	Schema                             = "gooo/autonomous-change-proposal-predecessor-selection/v2"
	ObservationSchema                  = "gooo/language-readiness-api-observation/v1"
	ResolutionSchema                   = "gooo/autonomous-change-proposal-predecessor-resolution/v1"
	ReasonSelected                     = "PROPOSAL_PREDECESSOR_SELECTED"
	ReasonNotFound                     = "PROPOSAL_PREDECESSOR_NOT_FOUND"
	ReasonAmbiguous                    = "PROPOSAL_PREDECESSOR_AMBIGUOUS"
	ReasonEvidenceUnknown              = "PROPOSAL_PREDECESSOR_EVIDENCE_UNKNOWN"
	ReasonJobCardinality               = "PROPOSAL_SYNTHESIS_JOB_CARDINALITY"
	ReasonRunPaginationIncomplete      = "PROPOSAL_RUN_PAGINATION_INCOMPLETE"
	ReasonJobPaginationIncomplete      = "PROPOSAL_JOB_PAGINATION_INCOMPLETE"
	ReasonArtifactPaginationIncomplete = "PROPOSAL_ARTIFACT_PAGINATION_INCOMPLETE"
	ReasonAPIUnavailable               = "PROPOSAL_API_UNAVAILABLE"
	ReasonAPIPermissionDenied          = "PROPOSAL_API_PERMISSION_DENIED"
	ReasonResponseMalformed            = "PROPOSAL_RESPONSE_MALFORMED"
	ReasonArtifactPayloadUnavailable   = "PROPOSAL_ARTIFACT_PAYLOAD_UNAVAILABLE"
	ReasonRedirectOriginMismatch       = "PROPOSAL_REDIRECT_ORIGIN_MISMATCH"

	ResolutionConformancePass = "PASS"
	ResolutionFailClosed      = "FAIL_CLOSED"
	ResolutionLower           = "LOWER_RESOLUTION"
	ResolutionStage           = "proposal-predecessor"
	ResolutionStep            = "select-exact-predecessor"
)

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
	FailureReason     string
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

type ObservationEvidence struct {
	Schema           string `json:"schema"`
	CachePath        string `json:"cache_path"`
	CacheBytes       int    `json:"cache_bytes"`
	CacheDigest      string `json:"cache_digest"`
	ResponseTotal    int    `json:"response_total"`
	ResponseConsumed int    `json:"response_consumed"`
}

func (report Report) Ready() bool {
	return report.Decision == "SELECTED" && report.Reason == ReasonSelected
}

type ResolutionReceipt struct {
	Schema              string              `json:"schema"`
	Repository          string              `json:"repository"`
	CurrentHeadSHA      string              `json:"current_head_sha"`
	PredecessorSHA      string              `json:"predecessor_sha"`
	Conformance         string              `json:"conformance"`
	Decision            string              `json:"decision"`
	Reason              string              `json:"reason"`
	Resolution          string              `json:"resolution"`
	Stage               string              `json:"stage"`
	Step                string              `json:"step"`
	PromotionAuthority  bool                `json:"promotion_authority"`
	ReadinessDelta      *int                `json:"readiness_delta"`
	Selection           *Report             `json:"selection,omitempty"`
	ObservationEvidence ObservationEvidence `json:"observation_evidence"`
	ReportDigest        string              `json:"report_digest"`
}
