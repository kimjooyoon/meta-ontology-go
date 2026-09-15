package publicworkflowlineage

const (
	PolicySchema              = "gooo/public-workflow-lineage-policy/v1"
	ReportSchema              = "gooo/public-workflow-lineage-report/v1"
	ObservationSchema         = "gooo/public-workflow-lineage-observation/v1"
	EvaluatorSchema           = "gooo/public-workflow-lineage-evaluator/v1"
	PolicyName                = "EXACT_SUBJECT_WORKFLOW_LINEAGE_CONTRACT"
	DecisionClosed            = "CLOSED"
	DecisionUnknown           = "UNKNOWN"
	DecisionRefuted           = "REFUTED"
	StateExact                = "EXACT"
	StateStale                = "STALE"
	StateDirectMissing        = "DIRECT_MISSING"
	StateMismatch             = "MISMATCH"
	StateTampered             = "TAMPERED"
	StateCurrentDevFallback   = "CURRENT_DEV_FALLBACK"
	UnknownClassStale         = "STALE"
	UnknownClassDirectMissing = "DIRECT_MISSING"
	ProvenanceExact           = "EXACT"
	ProvenanceUnknown         = "UNKNOWN"
	ProvenanceRefuted         = "REFUTED"
	RefStateValue             = "VALUE"
	RefStateNull              = "NULL"
	RefStateEmpty             = "EMPTY"
	RefStateMissing           = "MISSING"
	CaseCount                 = 9
	ClosedCaseCount           = 3
	UnknownCaseCount          = 3
	RefutedCaseCount          = 3
	LineageEdgeCount          = 6
	SourceReceiptCount        = 2
	ConsumerReceiptCount      = 2
	EvidenceArtifactCount     = 21
	ObservationAllowed        = "ELIGIBLE"
	ObservationDenied         = "INELIGIBLE"
	ReadOnlyPermission        = "READ_ONLY"
	ExactSuccessReuse         = "EXACT_SUCCESS_ONLY"
	NoPromotionPermission     = "NONE"
)

type CaseSpec struct {
	ID             string `json:"id"`
	Decision       string `json:"decision"`
	LineageState   string `json:"lineage_state"`
	UnknownClass   string `json:"unknown_class,omitempty"`
	SourceSubject  string `json:"source_subject_sha"`
	SourceRunID    int64  `json:"source_run_id"`
	SourceRefState string `json:"source_ref_state,omitempty"`
}

type Policy struct {
	Schema                  string         `json:"schema"`
	EvaluatorSchema         string         `json:"evaluator_schema"`
	SourceDigest            string         `json:"source_digest"`
	SemanticDigest          string         `json:"semantic_digest"`
	EvaluatorDigest         string         `json:"evaluator_digest"`
	Package                 string         `json:"package"`
	Namespace               string         `json:"namespace"`
	Activity                string         `json:"activity"`
	Name                    string         `json:"name"`
	Repository              string         `json:"repository"`
	SourceWorkflow          string         `json:"source_workflow"`
	ConsumerWorkflow        string         `json:"consumer_workflow"`
	SourceIdentity          []string       `json:"source_identity"`
	SourceIdentityPriority  []string       `json:"source_identity_priority"`
	SourceSecondaryFields   []string       `json:"source_secondary_fields"`
	SourceAPIKey            string         `json:"source_api_key"`
	ArtifactIdentityFields  []string       `json:"artifact_identity_fields"`
	ArtifactSubjectBinding  string         `json:"artifact_subject_binding"`
	ConsumerIdentity        []string       `json:"consumer_identity"`
	ProvenanceState         string         `json:"provenance_state"`
	LineageStates           []string       `json:"lineage_states"`
	CausalFields            []string       `json:"causal_fields"`
	LineageEdges            []string       `json:"lineage_edges"`
	RefutedDominatesUnknown bool           `json:"refuted_dominates_unknown"`
	ReadOnlyPermissions     Permissions    `json:"read_only_permissions"`
	Metrics                 map[string]int `json:"metrics"`
	Cases                   []CaseSpec     `json:"cases"`
}

// ReadOnlyPermissions binds observation eligibility to the existing semantic
// activities without granting the observation projection reuse or promotion
// authority.
type Permissions struct {
	WorkflowWindow      string `json:"workflow_window"`
	VerificationRuntime string `json:"verification_runtime"`
	EvidenceReuse       string `json:"evidence_reuse"`
	Promotion           string `json:"promotion"`
}

type ReadOnlyPermissions = Permissions

type Trigger struct {
	SourceWorkflow      string `json:"source_workflow"`
	SourceRunID         int64  `json:"source_run_id"`
	SourceRunAttempt    int64  `json:"source_run_attempt"`
	SourceSubjectSHA    string `json:"source_subject_sha"`
	SourceRef           string `json:"source_ref"`
	SourceRefState      string `json:"source_ref_state,omitempty"`
	SourceHeadBranch    string `json:"source_head_branch,omitempty"`
	SourceEvent         string `json:"source_event"`
	SourceRepository    string `json:"source_repository"`
	ConsumerWorkflow    string `json:"consumer_workflow"`
	ConsumerRunID       int64  `json:"consumer_run_id"`
	ConsumerRunAttempt  int64  `json:"consumer_run_attempt"`
	ConsumerSubjectSHA  string `json:"consumer_subject_sha"`
	ConsumerRef         string `json:"consumer_ref"`
	CandidateSubjectSHA string `json:"candidate_subject_sha,omitempty"`
}

type SourceRun struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Workflow          string `json:"workflow_name"`
	WorkflowPath      string `json:"workflow_path"`
	WorkflowID        int64  `json:"workflow_id"`
	Event             string `json:"event"`
	Ref               string `json:"ref"`
	RefState          string `json:"ref_state"`
	HeadBranch        string `json:"head_branch"`
	HeadSHA           string `json:"head_sha"`
	Repository        string `json:"head_repository"`
	APIRepositoryName string `json:"api_repository"`
	APIQueryRunID     int64  `json:"api_query_run_id"`
	ResolvedBy        string `json:"resolved_by"`
	Status            string `json:"status"`
	Conclusion        string `json:"conclusion"`
	RunAttempt        int64  `json:"run_attempt"`
}

type Artifact struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Digest         string `json:"digest"`
	PayloadDigest  string `json:"payload_digest,omitempty"`
	Size           int64  `json:"size_in_bytes"`
	RunID          int64  `json:"run_id"`
	RunAttempt     int64  `json:"run_attempt"`
	SubjectSHA     string `json:"subject_sha,omitempty"`
	SubjectBinding string `json:"subject_binding,omitempty"`
	Expired        bool   `json:"expired"`
}

type ArtifactIndex struct {
	LookupStatus string     `json:"lookup_status"`
	Artifacts    []Artifact `json:"artifacts"`
}

type Input struct {
	Trigger                        Trigger
	Source                         SourceRun
	Artifacts                      ArtifactIndex
	ExpectedArtifactName           string
	ArtifactStatus                 string
	ExpectedDigest                 string
	ExpectedRepository             string
	ExpectedWorkflow               string
	ExpectedSourceAPIKey           string
	ExpectedArtifactSubjectBinding string
}

type CausalUnknown struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Evaluation struct {
	Decision            string         `json:"decision"`
	LineageState        string         `json:"lineage_state"`
	Reason              string         `json:"reason"`
	Unknown             *CausalUnknown `json:"unknown,omitempty"`
	ExactSubjectBinding bool           `json:"exact_subject_binding"`
	MismatchDetected    bool           `json:"mismatch_detected"`
	FallbackAttempted   bool           `json:"fallback_attempted"`
	FallbackRejected    bool           `json:"fallback_rejected"`
	ArtifactResolved    bool           `json:"artifact_resolved"`
	ProvenanceState     string         `json:"provenance_state"`
	ProductFailureKept  bool           `json:"product_failure_kept"`
	ArtifactIdentity    string         `json:"artifact_identity,omitempty"`
}

type ReadOnlyObservationEvaluation struct {
	Schema                       string         `json:"schema"`
	Eligibility                  string         `json:"eligibility"`
	Decision                     string         `json:"decision"`
	LineageState                 string         `json:"lineage_state"`
	Reason                       string         `json:"reason"`
	ExactSourceIdentity          bool           `json:"exact_source_identity"`
	TimingObservationEligible    bool           `json:"timing_observation_eligible"`
	OperationObservationEligible bool           `json:"operation_observation_eligible"`
	EvidenceReuseAllowed         bool           `json:"evidence_reuse_allowed"`
	PromotionAllowed             bool           `json:"promotion_allowed"`
	SourceFailureKept            bool           `json:"source_failure_kept"`
	Unknown                      *CausalUnknown `json:"unknown,omitempty"`
}
