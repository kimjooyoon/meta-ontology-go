package main

const selectiveCIShadowUsage = "usage: gooo selective-ci shadow --base-snapshot FILE --head-snapshot FILE --plan-input FILE --evidence-input FILE --lane-input FILE"
const selectiveCIShadowSchemaVersion = "gooo/selective-ci-shadow/v1"

type selectiveCIShadowOptions struct {
	baseSnapshot  string
	headSnapshot  string
	planInput     string
	evidenceInput string
	laneInput     string
}
type shadowInputFiles struct {
	baseSnapshot  []byte
	headSnapshot  []byte
	planInput     []byte
	evidenceInput []byte
	laneInput     []byte
}
type shadowCommandSpec struct {
	ID   string   `json:"id"`
	Argv []string `json:"argv"`
}
type shadowResourceReceipt struct {
	CommandID    string `json:"command_id"`
	CPUWorkUnits uint64 `json:"cpu_work_units"`
	MemoryBytes  uint64 `json:"memory_bytes"`
}
type shadowLaneReceipt struct {
	Decision       string `json:"decision"`
	Reason         string `json:"reason"`
	RegistryDigest string `json:"registry_digest"`
	BaseSHA        string `json:"base_sha"`
	LaneHeadSHA    string `json:"lane_head_sha"`
	LaneID         string `json:"lane_id"`
}

// selectiveCIShadowOutput is intentionally a receipt, not an execution plan.
// It contains argv as data only; no field authorizes or represents execution.
type selectiveCIShadowOutput struct {
	SchemaVersion       string                  `json:"schema_version"`
	Command             string                  `json:"command"`
	Status              string                  `json:"status"`
	Stage               string                  `json:"stage"`
	Component           string                  `json:"component"`
	Reason              string                  `json:"reason"`
	ExecutionAuthorized bool                    `json:"execution_authorized"`
	ShadowOnly          bool                    `json:"shadow_only"`
	BaseSourceDigest    string                  `json:"base_source_digest"`
	HeadSourceDigest    string                  `json:"head_source_digest"`
	BaseSemanticDigest  string                  `json:"base_semantic_digest"`
	HeadSemanticDigest  string                  `json:"head_semantic_digest"`
	RegistryDigest      string                  `json:"registry_digest"`
	PlanDigest          string                  `json:"plan_digest"`
	ProofStatus         string                  `json:"proof_status"`
	ProofCode           string                  `json:"proof_code"`
	ChangedSemanticIDs  []string                `json:"changed_semantic_ids"`
	SelectedCommands    []shadowCommandSpec     `json:"selected_commands"`
	SelectedGuards      []shadowCommandSpec     `json:"selected_guard_commands"`
	SelectedWorkIDs     []string                `json:"selected_work_ids"`
	ResourceReceipts    []shadowResourceReceipt `json:"resource_receipts"`
	Lane                shadowLaneReceipt       `json:"lane"`
	CanonicalDigest     string                  `json:"canonical_digest"`
}
