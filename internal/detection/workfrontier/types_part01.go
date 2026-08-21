package workfrontier

// SchemaVersion identifies the work-frontier interchange contract.
const SchemaVersion = "gooo/work-frontier/v1"

// Decision is the selector's authorization state.
type Decision string

const (
	DecisionPass    Decision = "PASS"
	DecisionBlocked Decision = "BLOCKED"
	DecisionUnknown Decision = "UNKNOWN"
	StatusPass               = "PASS"
	StatusBlocked            = "BLOCKED"
)

// Pressure is a registry entry identified only by a stable ID.
type Pressure struct {
	StableID string `json:"stable_id"`

	// ID is a source-compatible convenience for callers that use ID for stable
	// identities. It is not part of the wire schema; StableID is authoritative.
	ID string `json:"-"`

	stableIDPresent bool
	fromJSON        bool
}

// ObligationState records the authoritative state of one obligation. Any
// non-empty state other than PASS is not complete and cannot satisfy a
// prerequisite.
type ObligationState struct {
	ObligationID string `json:"obligation_id"`
	Status       string `json:"status"`

	// ID is accepted for direct Go construction only. The JSON field is
	// obligation_id so identity cannot be inferred from a display name.
	ID string `json:"-"`

	obligationIDPresent bool
	statusPresent       bool
	fromJSON            bool
}

// RepairPath is one declared, finite route for repairing an obligation.
type RepairPath struct {
	StableID                  string   `json:"stable_id"`
	WorkID                    string   `json:"work_id"`
	ObligationID              string   `json:"obligation_id"`
	PrerequisiteObligationIDs []string `json:"prerequisite_obligation_ids"`
	ReadSet                   []string `json:"read_set"`
	WriteSet                  []string `json:"write_set"`
	RequiredPressureIDs       []string `json:"required_pressure_ids"`
	PolicyPriority            uint32   `json:"policy_priority"`
	CPUCoreNSUpperBound       uint64   `json:"cpu_core_ns_upper_bound"`

	// ID is accepted for direct Go construction only; StableID is the wire
	// identity. This alias avoids making display names part of selection.
	ID string `json:"-"`

	stableIDPresent                  bool
	obligationIDPresent              bool
	prerequisiteObligationIDsPresent bool
	readSetPresent                   bool
	writeSetPresent                  bool
	requiredPressureIDsPresent       bool
	policyPriorityPresent            bool
	cpuCoreNSUpperBoundPresent       bool
	fromJSON                         bool
}
