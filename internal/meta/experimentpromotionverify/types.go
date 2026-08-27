package experimentpromotionverify

const (
	ContractSchema      = "gooo/experiment-promotion-contract/v2"
	ReportSchema        = "gooo/experiment-promotion-report/v2"
	ObservationSchema   = "gooo/experiment-observation-receipt/v2"
	SourcePath          = "examples/experiment-promotion/main.gooo"
	ExperimentCount     = 30
	GateCount           = 5
	GateSlotCount       = 150
	CounterexampleCount = 9
	PortfolioScope      = "GOOO_META_EXPERIMENT_PROMOTION_LEDGER"
	StatusProven        = "PROVEN"
	StatusOpen          = "OPEN"
	StatusUnknown       = "UNKNOWN"
	StatusRefuted       = "REFUTED"
	ClaimOpen           = "OPEN"
	ClaimDischarged     = "DISCHARGED"
	ClaimRefuted        = "REFUTED"
	ObservationSuccess  = "success"
	ObservationProgress = "in_progress"
	ObservationQueued   = "queued"
	ObservationFailure  = "failure"
	ObservationCanceled = "cancelled"
	EvidenceCurrent     = "CURRENT_EVIDENCE"
	EvidenceHistorical  = "HISTORICAL_FIXTURE"
	EvidenceUnknown     = "UNKNOWN"
	ConsumerPackage     = "github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentpromotionverify"
	ConsumerProcedureID = "experimentpromotionverify.reconstruct/v2"
	ConsumerSourcePath  = "internal/meta/experimentpromotionverify"
)

var GateIDs = []string{"source-bound", "semantic-causality", "independent-consumer", "persistent-claim-transition", "exact-actions"}

type ExperimentIdentity struct {
	ID           string `json:"id"`
	PRNumber     int    `json:"pr_number"`
	Topic        string `json:"topic"`
	ClaimAddress string `json:"claim_address"`
}
type Contract struct {
	Schema                string               `json:"schema"`
	Version               int                  `json:"version"`
	SourcePath            string               `json:"source_path"`
	Experiments           []ExperimentIdentity `json:"experiments"`
	Gates                 []string             `json:"gates"`
	ExperimentDenominator int                  `json:"experiment_denominator"`
	GateSlotDenominator   int                  `json:"gate_slot_denominator"`
	RequiredReceiptFields []string             `json:"required_receipt_fields"`
	NotClaimed            []string             `json:"not_claimed"`
}
type Input struct {
	SubjectSHA         string
	SourceRaw          []byte
	ObservationRaw     []byte
	Contract           Contract
	Report             Report
	RepositorySnapshot RepositorySnapshot
}
type RepositorySnapshot struct {
	BeforeRaw       []byte   `json:"before_raw"`
	BeforeDigest    string   `json:"before_digest"`
	AfterRaw        []byte   `json:"after_raw"`
	AfterDigest     string   `json:"after_digest"`
	ChangedPaths    int      `json:"changed_paths"`
	ChangedPathList []string `json:"changed_path_list"`
}
type ObservationBundle struct {
	Schema           string                      `json:"schema"`
	BundleID         string                      `json:"bundle_id"`
	ObservationClass string                      `json:"observation_class"`
	Receipts         []ObservationReceipt        `json:"receipts"`
	PriorLedger      LedgerObservation           `json:"prior_ledger"`
	Counterexamples  []CounterexampleObservation `json:"counterexamples"`
}
type LedgerObservation struct {
	Path       string `json:"path"`
	Raw        []byte `json:"raw"`
	Digest     string `json:"digest"`
	EntryCount int    `json:"entry_count"`
	LastDigest string `json:"last_digest"`
}
type CounterexampleObservation struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Raw    []byte `json:"raw"`
	Digest string `json:"digest"`
}
type ObservationReceipt struct {
	Schema                   string                `json:"schema"`
	ObservationID            string                `json:"observation_id"`
	ExperimentID             string                `json:"experiment_id"`
	GateID                   string                `json:"gate_id"`
	PRNumber                 int                   `json:"pr_number"`
	ClaimAddress             string                `json:"claim_address"`
	EvidenceClass            string                `json:"evidence_class"`
	HeadSHA                  string                `json:"head_sha"`
	SourceRawDigest          string                `json:"source_raw_digest"`
	SourceSemanticDigest     string                `json:"source_semantic_digest"`
	ProducerID               string                `json:"producer_id"`
	ConsumerPackage          string                `json:"consumer_package"`
	ConsumerImports          []string              `json:"consumer_imports"`
	ClaimClass               string                `json:"claim_class"`
	ClaimTransitionDigest    string                `json:"claim_transition_digest"`
	ProcedureID              string                `json:"procedure_id"`
	ProcedureSourcePath      string                `json:"procedure_source_path"`
	ProcedureSourceBytes     []byte                `json:"procedure_source_bytes"`
	ProcedureSourceDigest    string                `json:"procedure_source_digest"`
	ProcedureAlgorithmID     string                `json:"procedure_algorithm_id"`
	ProcedureAlgorithmDigest string                `json:"procedure_algorithm_digest"`
	TargetAddress            string                `json:"target_address"`
	Actions                  ActionsObservation    `json:"actions"`
	Artifact                 ArtifactObservation   `json:"artifact"`
	SemanticIntervention     *SemanticIntervention `json:"semantic_intervention,omitempty"`
}
type ActionsObservation struct {
	Repository     string `json:"repository"`
	PRNumber       int    `json:"pr_number"`
	HeadSHA        string `json:"head_sha"`
	WorkflowID     string `json:"workflow_id"`
	WorkflowName   string `json:"workflow_name"`
	RunID          int64  `json:"run_id"`
	JobID          int64  `json:"job_id"`
	RunURL         string `json:"run_url"`
	JobURL         string `json:"job_url"`
	Conclusion     string `json:"conclusion"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactName   string `json:"artifact_name"`
	ArtifactDigest string `json:"artifact_digest"`
	Raw            []byte `json:"raw"`
	RawDigest      string `json:"raw_digest"`
}
type ArtifactObservation struct {
	Bytes         int    `json:"bytes"`
	Path          string `json:"path"`
	Digest        string `json:"digest"`
	TargetAddress string `json:"target_address"`
	ArtifactID    int64  `json:"artifact_id"`
	ArtifactName  string `json:"artifact_name"`
	Raw           []byte `json:"raw"`
}
type SemanticIntervention struct {
	BaselineSourceRaw      []byte `json:"baseline_source_raw"`
	SemanticSourceRaw      []byte `json:"semantic_source_raw"`
	CommentSourceRaw       []byte `json:"comment_source_raw"`
	BaselineRawDigest      string `json:"baseline_raw_digest"`
	BaselineSemanticDigest string `json:"baseline_semantic_digest"`
	SemanticRawDigest      string `json:"semantic_raw_digest"`
	SemanticSemanticDigest string `json:"semantic_semantic_digest"`
	CommentRawDigest       string `json:"comment_raw_digest"`
	CommentSemanticDigest  string `json:"comment_semantic_digest"`
	ContractedOutputDigest string `json:"contracted_output_digest"`
	DecisionDigest         string `json:"decision_digest"`
	ClaimTransitionDigest  string `json:"claim_transition_digest"`
}
type SourceProjection struct {
	Path           string               `json:"path"`
	RawDigest      string               `json:"raw_digest"`
	SemanticDigest string               `json:"semantic_digest"`
	Experiments    []ExperimentIdentity `json:"experiments"`
	Gates          []string             `json:"gates"`
	Exact          bool                 `json:"exact"`
}
type ClaimTransition struct {
	ExperimentID string `json:"experiment_id"`
	GateID       string `json:"gate_id"`
	From         string `json:"from"`
	To           string `json:"to"`
	Stage        string `json:"stage"`
	Step         string `json:"step"`
	Reason       string `json:"reason"`
	Digest       string `json:"digest"`
}
type GateResult struct {
	ExperimentID    string              `json:"experiment_id"`
	GateID          string              `json:"gate_id"`
	Status          string              `json:"status"`
	PromotionStatus string              `json:"promotion_status"`
	EvidenceClass   string              `json:"evidence_class"`
	ObservationID   string              `json:"observation_id,omitempty"`
	Stage           string              `json:"stage"`
	Step            string              `json:"step"`
	Reason          string              `json:"reason"`
	ClaimTransition ClaimTransition     `json:"claim_transition"`
	Receipt         *ObservationReceipt `json:"receipt,omitempty"`
}
type ExperimentResult struct {
	ExperimentID   string       `json:"experiment_id"`
	Status         string       `json:"status"`
	EvidenceStatus string       `json:"evidence_status"`
	EvidenceClass  string       `json:"evidence_class"`
	Stage          string       `json:"stage"`
	Step           string       `json:"step"`
	Reason         string       `json:"reason"`
	Gates          []GateResult `json:"gates"`
}
type ClaimLedgerEntry struct {
	Sequence       int    `json:"sequence"`
	ExperimentID   string `json:"experiment_id"`
	GateID         string `json:"gate_id"`
	PriorState     string `json:"prior_state"`
	NextState      string `json:"next_state"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	PreviousDigest string `json:"previous_digest"`
	Digest         string `json:"digest"`
}
type EmittedClaim struct {
	ExperimentID  string `json:"experiment_id"`
	GateID        string `json:"gate_id"`
	Class         string `json:"class"`
	State         string `json:"state"`
	TargetAddress string `json:"target_address"`
}
type StateCounts struct {
	Proven  int `json:"PROVEN"`
	Open    int `json:"OPEN"`
	Unknown int `json:"UNKNOWN"`
	Refuted int `json:"REFUTED"`
}
type Summary struct {
	DeclaredExperimentsNumerator       int         `json:"declared_experiments_numerator"`
	DeclaredExperimentsDenominator     int         `json:"declared_experiments_denominator"`
	MaterializedClaimSlotsNumerator    int         `json:"materialized_claim_slots_numerator"`
	MaterializedClaimSlotsDenominator  int         `json:"materialized_claim_slots_denominator"`
	ExperimentStates                   StateCounts `json:"experiment_states"`
	GateStates                         StateCounts `json:"gate_states"`
	FixtureExperimentStates            StateCounts `json:"fixture_experiment_states"`
	FixtureGateStates                  StateCounts `json:"fixture_gate_states"`
	ClaimTransitionsNumerator          int         `json:"claim_transitions_numerator"`
	ClaimTransitionsDenominator        int         `json:"claim_transitions_denominator"`
	RepositoryWrites                   int         `json:"repository_writes"`
	MutationAuthority                  bool        `json:"mutation_authority"`
	ForbiddenAggregates                int         `json:"forbidden_aggregates"`
	CounterexamplesDetectedNumerator   int         `json:"counterexamples_detected_numerator"`
	CounterexamplesDetectedDenominator int         `json:"counterexamples_detected_denominator"`
}
type Guardrail struct {
	ID                     string `json:"id"`
	Direction              string `json:"direction"`
	Observed               int    `json:"observed"`
	AllowedMax             int    `json:"allowed_max"`
	ConformanceNumerator   int    `json:"conformance_numerator"`
	ConformanceDenominator int    `json:"conformance_denominator"`
	Conforms               bool   `json:"conforms"`
}
type CounterexampleResult struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Detected bool   `json:"detected"`
	Stage    string `json:"stage"`
	Step     string `json:"step"`
	Reason   string `json:"reason"`
}
type Report struct {
	Schema             string                 `json:"schema"`
	Scope              string                 `json:"scope"`
	SubjectSHA         string                 `json:"subject_sha"`
	SourceProjection   SourceProjection       `json:"source_projection"`
	ObservationDigest  string                 `json:"observation_digest"`
	PriorLedger        LedgerObservation      `json:"prior_ledger"`
	Experiments        []ExperimentResult     `json:"experiments"`
	ClaimLedger        []ClaimLedgerEntry     `json:"claim_ledger"`
	EmittedClaims      []EmittedClaim         `json:"emitted_claims"`
	Summary            Summary                `json:"summary"`
	Guardrails         []Guardrail            `json:"guardrails"`
	Counterexamples    []CounterexampleResult `json:"counterexamples"`
	AggregateMetrics   []string               `json:"aggregate_metrics"`
	NotClaimed         []string               `json:"not_claimed"`
	RepositoryWrites   int                    `json:"repository_writes"`
	MutationAuthority  bool                   `json:"mutation_authority"`
	RepositorySnapshot RepositorySnapshot     `json:"repository_snapshot"`
	Digest             string                 `json:"digest"`
}
type ReplayCheck struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Stage    string `json:"stage"`
	Step     string `json:"step"`
	Reason   string `json:"reason"`
	Expected string `json:"expected"`
	Observed string `json:"observed"`
}
type Verification struct {
	Schema            string                 `json:"schema"`
	SubjectSHA        string                 `json:"subject_sha"`
	Decision          string                 `json:"decision"`
	Resolution        string                 `json:"resolution"`
	Reason            string                 `json:"reason"`
	SourceProjection  SourceProjection       `json:"source_projection"`
	Checks            []ReplayCheck          `json:"checks"`
	Summary           Summary                `json:"summary"`
	Guardrails        []Guardrail            `json:"guardrails"`
	Counterexamples   []CounterexampleResult `json:"counterexamples"`
	AggregateMetrics  []string               `json:"aggregate_metrics"`
	NotClaimed        []string               `json:"not_claimed"`
	RepositoryWrites  int                    `json:"repository_writes"`
	MutationAuthority bool                   `json:"mutation_authority"`
	Digest            string                 `json:"digest"`
}
