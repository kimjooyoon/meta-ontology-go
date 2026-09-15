package shadow

type productionCommand struct {
	ID   string   `json:"id"`
	Argv []string `json:"argv"`
}

type productionResourceReceipt struct {
	CommandID    string `json:"command_id"`
	CPUWorkUnits uint64 `json:"cpu_work_units"`
	MemoryBytes  uint64 `json:"memory_bytes"`
}

type productionLane struct {
	Decision       string `json:"decision"`
	Reason         string `json:"reason"`
	RegistryDigest string `json:"registry_digest"`
	BaseSHA        string `json:"base_sha"`
	LaneHeadSHA    string `json:"lane_head_sha"`
	LaneID         string `json:"lane_id"`
}

type productionOutput struct {
	SchemaVersion       string                      `json:"schema_version"`
	Command             string                      `json:"command"`
	Status              string                      `json:"status"`
	Stage               string                      `json:"stage"`
	Component           string                      `json:"component"`
	Reason              string                      `json:"reason"`
	ExecutionAuthorized bool                        `json:"execution_authorized"`
	ShadowOnly          bool                        `json:"shadow_only"`
	BaseSourceDigest    string                      `json:"base_source_digest"`
	HeadSourceDigest    string                      `json:"head_source_digest"`
	BaseSemanticDigest  string                      `json:"base_semantic_digest"`
	HeadSemanticDigest  string                      `json:"head_semantic_digest"`
	RegistryDigest      string                      `json:"registry_digest"`
	PlanDigest          string                      `json:"plan_digest"`
	ProofStatus         string                      `json:"proof_status"`
	ProofCode           string                      `json:"proof_code"`
	ChangedSemanticIDs  []string                    `json:"changed_semantic_ids"`
	SelectedCommands    []productionCommand         `json:"selected_commands"`
	SelectedGuards      []productionCommand         `json:"selected_guard_commands"`
	SelectedWorkIDs     []string                    `json:"selected_work_ids"`
	ResourceReceipts    []productionResourceReceipt `json:"resource_receipts"`
	Lane                productionLane              `json:"lane"`
	CanonicalDigest     string                      `json:"canonical_digest"`
}
