package publicworkflowlineage

const (
	PolicySchema              = "gooo/public-workflow-lineage-policy/v1"
	ReportSchema              = "gooo/public-workflow-lineage-report/v1"
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
	CaseCount                 = 6
	ClosedCaseCount           = 2
	UnknownCaseCount          = 2
	RefutedCaseCount          = 2
	LineageEdgeCount          = 6
	SourceReceiptCount        = 2
	ConsumerReceiptCount      = 2
	EvidenceArtifactCount     = 18
)

type CaseSpec struct {
	ID            string `json:"id"`
	Decision      string `json:"decision"`
	LineageState  string `json:"lineage_state"`
	UnknownClass  string `json:"unknown_class,omitempty"`
	SourceSubject string `json:"source_subject_sha"`
	SourceRunID   int64  `json:"source_run_id"`
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
	SourceWorkflow          string         `json:"source_workflow"`
	ConsumerWorkflow        string         `json:"consumer_workflow"`
	SourceIdentity          []string       `json:"source_identity"`
	ArtifactIdentityFields  []string       `json:"artifact_identity_fields"`
	ConsumerIdentity        []string       `json:"consumer_identity"`
	LineageStates           []string       `json:"lineage_states"`
	CausalFields            []string       `json:"causal_fields"`
	LineageEdges            []string       `json:"lineage_edges"`
	RefutedDominatesUnknown bool           `json:"refuted_dominates_unknown"`
	Metrics                 map[string]int `json:"metrics"`
	Cases                   []CaseSpec     `json:"cases"`
}

type Trigger struct {
	SourceWorkflow      string `json:"source_workflow"`
	SourceRunID         int64  `json:"source_run_id"`
	SourceRunAttempt    int64  `json:"source_run_attempt"`
	SourceSubjectSHA    string `json:"source_subject_sha"`
	SourceRef           string `json:"source_ref"`
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
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Workflow   string `json:"workflow_name"`
	Event      string `json:"event"`
	Ref        string `json:"ref"`
	HeadBranch string `json:"head_branch"`
	HeadSHA    string `json:"head_sha"`
	Repository string `json:"head_repository"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	RunAttempt int64  `json:"run_attempt"`
}

type Artifact struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Digest        string `json:"digest"`
	PayloadDigest string `json:"payload_digest,omitempty"`
	Size          int64  `json:"size_in_bytes"`
	RunID         int64  `json:"run_id"`
	SubjectSHA    string `json:"subject_sha,omitempty"`
	Expired       bool   `json:"expired"`
}

type ArtifactIndex struct {
	LookupStatus string     `json:"lookup_status"`
	Artifacts    []Artifact `json:"artifacts"`
}

type Input struct {
	Trigger              Trigger
	Source               SourceRun
	Artifacts            ArtifactIndex
	ExpectedArtifactName string
	ArtifactStatus       string
	ExpectedDigest       string
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
	ProductFailureKept  bool           `json:"product_failure_kept"`
	ArtifactIdentity    string         `json:"artifact_identity,omitempty"`
}
