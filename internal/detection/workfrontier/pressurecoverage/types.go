// Package pressurecoverage defines a strict canonical pressure-input envelope.
package pressurecoverage

const (
	SchemaVersion        = "gooo/workfrontier-pressure-coverage/v1"
	LanguageFloor uint64 = 2
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
}
