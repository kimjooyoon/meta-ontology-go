package proposalpredecessor

const (
	Schema                             = "gooo/autonomous-change-proposal-predecessor-selection/v2"
	ObservationSchema                  = "gooo/language-readiness-api-observation/v1"
	ObservationMemberPath              = "language-readiness-proposal-observation.json"
	ObservationRole                    = "proposal-predecessor-raw-observation"
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
	ReasonRouteUnknown                 = "PROPOSAL_PREDECESSOR_ROUTE_UNKNOWN"
	ReasonRouteContradiction           = "PROPOSAL_PREDECESSOR_ROUTE_CONTRADICTION"

	ResolutionConformancePass = "PASS"
	ResolutionFailClosed      = "FAIL_CLOSED"
	ResolutionLower           = "LOWER_RESOLUTION"
	ResolutionStage           = "proposal-predecessor"
	ResolutionStep            = "select-exact-predecessor"
	DecisionClosed            = "CLOSED"
	DecisionUnknown           = "UNKNOWN"
	DecisionRefuted           = "REFUTED"
	ResolutionExact           = "EXACT"
	RouteDev                  = "dev"
	RouteMain                 = "main"
	UnknownClassRoute         = "ROUTE_IDENTITY_MISSING"
	UnknownClassAmbiguous     = "MULTIPLE_EXACT_ROUTE_CANDIDATES"
	UnknownClassMissing       = "DIRECT_MISSING"
)

type Selected struct {
	RunID                  int64  `json:"run_id"`
	RunAttempt             int    `json:"run_attempt"`
	HeadSHA                string `json:"head_sha"`
	HeadBranch             string `json:"head_branch"`
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
	RequestedRoute    string
	ObservedRuns      int
	ExactRuns         int
	OtherRouteRuns    int
	RouteUnknownRuns  int
	Contradictions    int
	ObservedArtifacts int
	ExactArtifacts    int
	ObservedJobs      int
	ExactJobs         int
	Unresolved        int
	Candidates        []Candidate
	FailureReason     string
	pending           []githubRun
}

type Report struct {
	Schema                string      `json:"schema"`
	Repository            string      `json:"repository"`
	CurrentSubjectSHA     string      `json:"current_subject_sha"`
	PredecessorSHA        string      `json:"predecessor_sha"`
	RequestedRoute        string      `json:"requested_route"`
	Decision              string      `json:"decision"`
	Reason                string      `json:"reason"`
	ObservationDecision   string      `json:"observation_decision"`
	ObservationResolution string      `json:"observation_resolution"`
	Unknown               *Unknown    `json:"unknown,omitempty"`
	Selected              *Selected   `json:"selected,omitempty"`
	Summary               Summary     `json:"summary"`
	Indicators            []Indicator `json:"indicators"`
	Proofs                []Proof     `json:"proofs"`
	ReportDigest          string      `json:"report_digest"`
}

type ObservationEvidence struct {
	Schema           string `json:"schema"`
	CachePath        string `json:"cache_path"`
	CacheRole        string `json:"cache_role"`
	CacheBytes       int    `json:"cache_bytes"`
	CacheDigest      string `json:"cache_digest"`
	ResponseTotal    int    `json:"response_total"`
	ResponseConsumed int    `json:"response_consumed"`
}

type ObservationResponse struct {
	Kind       string `json:"kind"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
	Link       string `json:"link,omitempty"`
	Location   string `json:"location,omitempty"`
	Failure    string `json:"failure,omitempty"`
}

type ObservationCache struct {
	Schema    string                `json:"schema"`
	Responses []ObservationResponse `json:"responses"`
}

type Unknown struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func (report Report) Ready() bool {
	return report.Decision == "SELECTED" && report.Reason == ReasonSelected
}

type ResolutionReceipt struct {
	Schema                string              `json:"schema"`
	Repository            string              `json:"repository"`
	CurrentHeadSHA        string              `json:"current_head_sha"`
	PredecessorSHA        string              `json:"predecessor_sha"`
	RequestedRoute        string              `json:"requested_route"`
	Conformance           string              `json:"conformance"`
	Decision              string              `json:"decision"`
	Reason                string              `json:"reason"`
	Resolution            string              `json:"resolution"`
	Stage                 string              `json:"stage"`
	Step                  string              `json:"step"`
	ObservationDecision   string              `json:"observation_decision"`
	ObservationResolution string              `json:"observation_resolution"`
	Unknown               *Unknown            `json:"unknown,omitempty"`
	PromotionAuthority    bool                `json:"promotion_authority"`
	ReadinessDelta        *int                `json:"readiness_delta"`
	Selection             *Report             `json:"selection,omitempty"`
	ObservationEvidence   ObservationEvidence `json:"observation_evidence"`
	ReportDigest          string              `json:"report_digest"`
}
