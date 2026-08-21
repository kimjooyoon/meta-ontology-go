package resourcevector

const (
	SchemaV1       = "gooo/selective-ci-resource-vector/v1"
	CorpusSchemaV1 = "gooo/selective-ci-resource-vector-corpus/v1"
)

type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionUnknown    Decision = "UNKNOWN"
	DecisionFailClosed Decision = "FAIL_CLOSED"
)
const (
	Pass       = DecisionPass
	Unknown    = DecisionUnknown
	FailClosed = DecisionFailClosed
)

type Reason string

const (
	ReasonNone                     Reason = "NONE"
	ReasonMissingInput             Reason = "MISSING_INPUT"
	ReasonMissingResource          Reason = "MISSING_RESOURCE"
	ReasonMissingPROV              Reason = "MISSING_PROV"
	ReasonInvalidPath              Reason = "INVALID_PATH"
	ReasonDuplicateID              Reason = "DUPLICATE_ID"
	ReasonDuplicateRecord          Reason = "DUPLICATE_PROV_RECORD"
	ReasonDuplicatePath            Reason = "DUPLICATE_PATH"
	ReasonDanglingID               Reason = "DANGLING_ID"
	ReasonSelectionInvalid         Reason = "INVALID_SELECTION"
	ReasonInvalidPressure          Reason = "INVALID_PRESSURE"
	ReasonOverflow                 Reason = "COUNT_OVERFLOW"
	ReasonResourceLimitExceeded    Reason = "RESOURCE_LIMIT_EXCEEDED"
	ReasonCeilingExceeded          Reason = ReasonResourceLimitExceeded
	ReasonMissingAffectedBinding   Reason = "MISSING_AFFECTED_BINDING"
	ReasonDuplicateAffectedBinding Reason = "DUPLICATE_AFFECTED_BINDING"
	ReasonDanglingAffectedBinding  Reason = "DANGLING_AFFECTED_BINDING"
	ReasonClosureInvalid           Reason = "INVALID_PROV_CLOSURE"
	ReasonRootRelocationInvalid    Reason = "INVALID_ROOT_RELOCATION"
)

// PressureRecord is part of a canonical command record. Applicable is a
// pointer so an omitted field cannot be mistaken for false.
type PressureRecord struct {
	ID                  string `json:"id"`
	IndependenceGroupID string `json:"independence_group_id"`
	Applicable          *bool  `json:"applicable"`
}

// CommandRecord is the canonical command-side record. Resource dimensions
// are presence-aware: nil means the producer did not provide the field.
type CommandRecord struct {
	ID                string           `json:"id"`
	Path              string           `json:"path"`
	CPUCoreNS         *uint64          `json:"cpu_core_ns"`
	MemoryBytes       *uint64          `json:"memory_bytes"`
	PeakMemoryBytes   *uint64          `json:"peak_memory_bytes"`
	WorkUnits         *uint64          `json:"work_units"`
	Pressures         []PressureRecord `json:"pressures"`
	AffectedStableIDs []string         `json:"affected_stable_ids"`
}
