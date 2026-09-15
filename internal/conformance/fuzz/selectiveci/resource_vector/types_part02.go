package resourcevector

// PathRecord is the canonical provenance-side record. Record IDs are kept as
// an ordered list in the input, but the oracle counts them only after checking
// global uniqueness. No duplicate is silently collapsed.
type PathRecord struct {
	ID                 string   `json:"id"`
	Path               string   `json:"path"`
	CommandID          string   `json:"command_id"`
	RecordIDs          []string `json:"record_ids"`
	Finite             *bool    `json:"finite"`
	ClosureNumerator   *uint64  `json:"closure_numerator"`
	ClosureDenominator *uint64  `json:"closure_denominator"`
}

// CeilingSet contains an explicit ceiling for every replay dimension.
// Pointers distinguish an omitted ceiling from an intentional numeric value.
type CeilingSet struct {
	CPUCoreNS           *uint64 `json:"cpu_core_ns"`
	MemoryBytes         *uint64 `json:"memory_bytes"`
	PeakMemoryBytes     *uint64 `json:"peak_memory_bytes"`
	WorkUnits           *uint64 `json:"work_units"`
	AffectedStableIDs   *uint64 `json:"affected_stable_ids"`
	ApplicablePressures *uint64 `json:"applicable_pressures"`
	IndependentGroups   *uint64 `json:"independent_groups"`
	UniquePROVRecords   *uint64 `json:"unique_prov_records"`
	FinitePROVPaths     *uint64 `json:"finite_prov_paths"`
	ClosureNumerator    *uint64 `json:"closure_numerator"`
	ClosureDenominator  *uint64 `json:"closure_denominator"`
}
type ResourceCeilings struct {
	Selected CeilingSet `json:"selected"`
	Full     CeilingSet `json:"full"`
}

// Input contains only producer-side canonical records. Expected labels are
// held by CorpusCase and are intentionally not part of this value.
type Input struct {
	Schema             string           `json:"schema"`
	FixtureID          string           `json:"fixture_id"`
	Root               string           `json:"root"`
	Commands           []CommandRecord  `json:"commands"`
	Paths              []PathRecord     `json:"paths"`
	AffectedStableIDs  []string         `json:"affected_stable_ids"`
	SelectedCommandIDs []string         `json:"selected_command_ids"`
	FullCommandIDs     []string         `json:"full_command_ids"`
	Ceilings           ResourceCeilings `json:"ceilings"`
}

// Vector is a fully recomputed aggregate. A vector is present in Output only
// when all fields needed to compute it are present and finite.
type Vector struct {
	CPUCoreNS           uint64 `json:"cpu_core_ns"`
	MemoryBytes         uint64 `json:"memory_bytes"`
	PeakMemoryBytes     uint64 `json:"peak_memory_bytes"`
	WorkUnits           uint64 `json:"work_units"`
	AffectedStableIDs   uint64 `json:"affected_stable_ids"`
	ApplicablePressures uint64 `json:"applicable_pressures"`
	IndependentGroups   uint64 `json:"independent_groups"`
	UniquePROVRecords   uint64 `json:"unique_prov_records"`
	FinitePROVPaths     uint64 `json:"finite_prov_paths"`
	ClosureNumerator    uint64 `json:"closure_numerator"`
	ClosureDenominator  uint64 `json:"closure_denominator"`
}
