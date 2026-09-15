package workfrontier

// Capacity is the available CPU budget for one selection.
type Capacity struct {
	CPUCoreNS uint64 `json:"cpu_core_ns"`

	cpuCoreNSPresent bool
}

// Input is the complete digest-bound work-frontier snapshot.
type Input struct {
	SchemaVersion            string            `json:"schema_version"`
	SnapshotDigest           string            `json:"snapshot_digest"`
	PolicyDigest             string            `json:"policy_digest"`
	RegistryDigest           string            `json:"registry_digest"`
	MinimumSelectedPressures uint32            `json:"minimum_selected_pressures"`
	Capacity                 Capacity          `json:"capacity"`
	Pressures                []Pressure        `json:"pressures"`
	States                   []ObligationState `json:"states"`
	Paths                    []RepairPath      `json:"paths"`

	fromJSON bool
	present  inputPresence
}
type inputPresence struct {
	schemaVersion            bool
	snapshotDigest           bool
	policyDigest             bool
	registryDigest           bool
	minimumSelectedPressures bool
	capacity                 bool
	pressures                bool
	states                   bool
	paths                    bool
}

// Result is the deterministic frontier partition. Selected contains work IDs;
// the other sets contain stable path IDs. A shortfall is an UNKNOWN result.
type Result struct {
	Status            Decision `json:"status"`
	Selected          []string `json:"selected"`
	SelectedIDs       []string `json:"selected_ids"`
	WorkIDs           []string `json:"work_ids"`
	Unknown           []string `json:"unknown"`
	Blocked           []string `json:"blocked"`
	Shortfall         []string `json:"shortfall"`
	Quality           string   `json:"quality"`
	FullSuiteRequired bool     `json:"full_suite_required"`
}

// SelectionResult is an expressive alias for Result.
type SelectionResult = Result

func (p Pressure) stableID() string {
	if p.StableID != "" {
		return p.StableID
	}
	return p.ID
}
func (s ObligationState) obligationID() string {
	if s.ObligationID != "" {
		return s.ObligationID
	}
	return s.ID
}
func (p RepairPath) stableID() string {
	if p.StableID != "" {
		return p.StableID
	}
	return p.ID
}
