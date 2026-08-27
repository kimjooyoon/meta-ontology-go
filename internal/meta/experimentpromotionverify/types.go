package experimentpromotionverify

const (
	ReportSchema        = "gooo/experiment-promotion-report/v1"
	ObservationSchema   = "gooo/experiment-observation-receipt/v1"
	SourcePath          = "examples/experiment-promotion/main.gooo"
	ExperimentCount     = 30
	GateCount           = 5
	GateSlotCount       = 150
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
)

var GateIDs = []string{"source-bound", "semantic-causality", "independent-consumer", "persistent-claim-transition", "exact-actions"}

type Contract struct {
	Schema                string   `json:"schema"`
	Version               int      `json:"version"`
	SourcePath            string   `json:"source_path"`
	Experiments           []string `json:"experiments"`
	Gates                 []string `json:"gates"`
	ExperimentDenominator int      `json:"experiment_denominator"`
	GateSlotDenominator   int      `json:"gate_slot_denominator"`
	RequiredReceiptFields []string `json:"required_receipt_fields"`
	NotClaimed            []string `json:"not_claimed"`
}

type Input struct {
	SourceRaw          []byte
	ObservationRaw     []byte
	Contract           Contract
	Report             Report
	SubjectSHA         string
	RepositorySnapshot RepositorySnapshot
}

type RepositorySnapshot struct {
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
	ChangedPaths int    `json:"changed_paths"`
}

type ObservationBundle struct {
	Schema   string               `json:"schema"`
	BundleID string               `json:"bundle_id"`
	Receipts []ObservationReceipt `json:"receipts"`
}

type ObservationReceipt struct {
	Schema                string                `json:"schema"`
	ObservationID         string                `json:"observation_id"`
	ExperimentID          string                `json:"experiment_id"`
	GateID                string                `json:"gate_id"`
	PRNumber              int                   `json:"pr_number"`
	HeadSHA               string                `json:"head_sha"`
	SourceRawDigest       string                `json:"source_raw_digest"`
	SourceSemanticDigest  string                `json:"source_semantic_digest"`
	ProducerID            string                `json:"producer_id"`
	ConsumerPackage       string                `json:"consumer_package"`
	ConsumerImports       []string              `json:"consumer_imports"`
	ClaimTransitionDigest string                `json:"claim_transition_digest"`
	Actions               ActionsObservation    `json:"actions"`
	Artifact              ArtifactObservation   `json:"artifact"`
	SemanticIntervention  *SemanticIntervention `json:"semantic_intervention,omitempty"`
}

type ActionsObservation struct {
	RunURL     string `json:"run_url"`
	JobURL     string `json:"job_url"`
	Conclusion string `json:"conclusion"`
}

type ArtifactObservation struct {
	Bytes  int    `json:"bytes"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type SemanticIntervention struct {
	BaselineRawDigest          string `json:"baseline_raw_digest"`
	BaselineSemanticDigest     string `json:"baseline_semantic_digest"`
	InterventionRawDigest      string `json:"intervention_raw_digest"`
	InterventionSemanticDigest string `json:"intervention_semantic_digest"`
}

type SourceProjection struct {
	Path           string   `json:"path"`
	RawDigest      string   `json:"raw_digest"`
	SemanticDigest string   `json:"semantic_digest"`
	Experiments    []string `json:"experiments"`
	Gates          []string `json:"gates"`
	Exact          bool     `json:"exact"`
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
	ObservationID   string              `json:"observation_id,omitempty"`
	Stage           string              `json:"stage"`
	Step            string              `json:"step"`
	Reason          string              `json:"reason"`
	ClaimTransition ClaimTransition     `json:"claim_transition"`
	Receipt         *ObservationReceipt `json:"receipt,omitempty"`
}

type ExperimentResult struct {
	ExperimentID string       `json:"experiment_id"`
	Status       string       `json:"status"`
	Stage        string       `json:"stage"`
	Step         string       `json:"step"`
	Reason       string       `json:"reason"`
	Gates        []GateResult `json:"gates"`
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
	ExperimentID string `json:"experiment_id"`
	GateID       string `json:"gate_id"`
	Class        string `json:"class"`
	State        string `json:"state"`
}

type StateCounts struct {
	Proven  int `json:"PROVEN"`
	Open    int `json:"OPEN"`
	Unknown int `json:"UNKNOWN"`
	Refuted int `json:"REFUTED"`
}

type Summary struct {
	ExperimentsNumerator        int         `json:"experiments_numerator"`
	ExperimentsDenominator      int         `json:"experiments_denominator"`
	GateSlotsNumerator          int         `json:"gate_slots_numerator"`
	GateSlotsDenominator        int         `json:"gate_slots_denominator"`
	ExperimentStates            StateCounts `json:"experiment_states"`
	GateStates                  StateCounts `json:"gate_states"`
	ClaimTransitionsNumerator   int         `json:"claim_transitions_numerator"`
	ClaimTransitionsDenominator int         `json:"claim_transitions_denominator"`
	RepositoryWrites            int         `json:"repository_writes"`
	MutationAuthority           bool        `json:"mutation_authority"`
	ForbiddenAggregates         int         `json:"forbidden_aggregates"`
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

type Report struct {
	Schema             string             `json:"schema"`
	Scope              string             `json:"scope"`
	SubjectSHA         string             `json:"subject_sha"`
	SourceProjection   SourceProjection   `json:"source_projection"`
	ObservationDigest  string             `json:"observation_digest"`
	Experiments        []ExperimentResult `json:"experiments"`
	ClaimLedger        []ClaimLedgerEntry `json:"claim_ledger"`
	EmittedClaims      []EmittedClaim     `json:"emitted_claims"`
	Summary            Summary            `json:"summary"`
	Guardrails         []Guardrail        `json:"guardrails"`
	AggregateMetrics   []string           `json:"aggregate_metrics"`
	NotClaimed         []string           `json:"not_claimed"`
	RepositoryWrites   int                `json:"repository_writes"`
	MutationAuthority  bool               `json:"mutation_authority"`
	RepositorySnapshot RepositorySnapshot `json:"repository_snapshot"`
	Digest             string             `json:"digest"`
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
	Schema            string           `json:"schema"`
	SubjectSHA        string           `json:"subject_sha"`
	Decision          string           `json:"decision"`
	Resolution        string           `json:"resolution"`
	Reason            string           `json:"reason"`
	SourceProjection  SourceProjection `json:"source_projection"`
	Checks            []ReplayCheck    `json:"checks"`
	Summary           Summary          `json:"summary"`
	Guardrails        []Guardrail      `json:"guardrails"`
	AggregateMetrics  []string         `json:"aggregate_metrics"`
	NotClaimed        []string         `json:"not_claimed"`
	RepositoryWrites  int              `json:"repository_writes"`
	MutationAuthority bool             `json:"mutation_authority"`
	Digest            string           `json:"digest"`
}
