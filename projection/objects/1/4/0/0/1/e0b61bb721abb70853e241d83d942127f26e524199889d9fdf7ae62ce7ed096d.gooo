package pressureindependence

type ResourceCeilings struct {
	CPUCoreNS   uint64 `json:"cpu_core_ns"`
	MemoryBytes uint64 `json:"memory_bytes"`
	WorkUnits   uint64 `json:"work_units"`
	ProvRecords uint64 `json:"prov_records"`
	ProvPaths   uint64 `json:"prov_paths"`
}
type CostReceipt struct {
	CPUCoreNS   uint64 `json:"cpu_core_ns"`
	MemoryBytes uint64 `json:"memory_bytes"`
	WorkUnits   uint64 `json:"work_units"`
	ProvRecords uint64 `json:"prov_records"`
	ProvPaths   uint64 `json:"prov_paths"`
}

// Output is the complete oracle result. It contains no authorization field.
type Output struct {
	Schema                string      `json:"schema"`
	FixtureID             string      `json:"fixture_id"`
	InputDigest           string      `json:"input_digest"`
	SelectedIDs           []string    `json:"selected_ids"`
	UnselectedIDs         []string    `json:"unselected_ids"`
	UnknownIDs            []string    `json:"unknown_ids"`
	DistinctGroupCount    uint64      `json:"distinct_group_count"`
	Decision              Decision    `json:"decision"`
	Reason                Reason      `json:"reason"`
	FullSuiteRequired     bool        `json:"full_suite_required"`
	ProofValid            bool        `json:"proof_valid"`
	CostReceipt           CostReceipt `json:"cost_receipt"`
	CanonicalOutputDigest string      `json:"canonical_output_digest"`
	ReplayDigest          string      `json:"replay_digest"`
}
type Finding string

const (
	NoUniqueBenefit             Finding = "NO_UNIQUE_BENEFIT"
	UniqueBenefitNotEstablished Finding = "UNIQUE_BENEFIT_NOT_ESTABLISHED"
)

// BaselineResult is a fair typed-config/full-suite comparison result.
type BaselineResult struct {
	Decision     Decision
	Reason       Reason
	LocalizedIDs []string
	FullSuite    bool
	CostReceipt  CostReceipt
	WorkUnits    uint64
}
type Comparison struct {
	Oracle            Output
	Baseline          BaselineResult
	OutcomeMatch      bool
	ReasonMatch       bool
	LocalizationMatch bool
	ResearchWorkUnits uint64
	BaselineWorkUnits uint64
	ResearchBudgetOK  bool
	Finding           Finding
}
