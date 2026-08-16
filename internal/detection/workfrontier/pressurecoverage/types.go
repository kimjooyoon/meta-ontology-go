// Package pressurecoverage observes explicit independent-pressure coverage.
package pressurecoverage

const SchemaVersion = "gooo/workfrontier-pressure-coverage/v1"
const LanguageFloor uint64 = 2

type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionFailClosed Decision = "FAIL_CLOSED"
	DecisionUnknown    Decision = "UNKNOWN"
)

type Reason string

const (
	ReasonNone                      Reason = "NONE"
	ReasonRequiredInputMissing      Reason = "REQUIRED_INPUT_MISSING"
	ReasonInputAmbiguous            Reason = "INPUT_AMBIGUOUS"
	ReasonStaleDigest               Reason = "STALE_DIGEST"
	ReasonApplicabilityUnproven     Reason = "APPLICABILITY_UNPROVEN"
	ReasonCatalogMismatch           Reason = "CATALOG_MISMATCH"
	ReasonInvalidStableID           Reason = "INVALID_STABLE_ID"
	ReasonIndependentGroupShortfall Reason = "INDEPENDENT_GROUP_SHORTFALL"
	ReasonDuplicateID               Reason = "DUPLICATE_ID"
	ReasonDuplicatePressureID       Reason = "DUPLICATE_PRESSURE_ID"
	ReasonConflictingGroupBinding   Reason = "CONFLICTING_GROUP_BINDING"
	ReasonMalformedFinitePath       Reason = "PROV_PATH_MALFORMED"
)

type PressureRecord struct {
	PressureID          string `json:"pressure_id"`
	CategoryID          string `json:"category_id"`
	IndependenceGroupID string `json:"independence_group_id"`
	ApplicabilityRuleID string `json:"applicability_rule_id"`
}
type Input struct {
	Schema                  string           `json:"schema"`
	AuthoritySnapshotDigest string           `json:"authority_snapshot_digest"`
	PolicyDigest            string           `json:"policy_digest"`
	RegistryDigest          string           `json:"registry_digest"`
	ToolchainOptionsDigest  string           `json:"toolchain_options_digest"`
	RequestedK              uint64           `json:"requested_K"`
	MinimumIndependent      uint64           `json:"minimum_independent"`
	PressureRecords         []PressureRecord `json:"pressure_records"`
	RequiredPressureIDs     []string         `json:"required_pressure_ids"`
	FinitePathIDs           []string         `json:"finite_path_ids"`
	GuardIDs                []string         `json:"guard_ids"`
}
type Output struct {
	Schema                 string   `json:"schema"`
	InputDigest            string   `json:"input_digest"`
	RequiredPressureCount  uint64   `json:"required_pressure_count"`
	DistinctGroupCount     uint64   `json:"distinct_group_count"`
	SelectedIDs            []string `json:"selected_ids"`
	UnselectedIDs          []string `json:"unselected_ids"`
	UnknownIDs             []string `json:"unknown_ids"`
	DeterministicWorkUnits uint64   `json:"deterministic_work_units"`
	CPUCoreNS              uint64   `json:"cpu_core_ns"`
	MemoryBytes            uint64   `json:"memory_bytes"`
	ProvRecords            uint64   `json:"prov_records"`
	ProvPaths              uint64   `json:"prov_paths"`
	Decision               Decision `json:"decision"`
	Reason                 Reason   `json:"reason"`
	FullSuiteRequired      bool     `json:"full_suite_required"`
	OutputDigest           string   `json:"output_digest"`
	ReplayDigest           string   `json:"replay_digest"`
}
