package fullsoundness

type Outcome struct {
	CommandID    string        `json:"command_id"`
	Status       OutcomeStatus `json:"status"`
	FailureCode  string        `json:"failure_code"`
	OutputDigest string        `json:"output_digest"`
}
type ResourceReceipt struct {
	CommandID         string `json:"command_id"`
	SnapshotDigest    string `json:"snapshot_digest"`
	ToolchainDigest   string `json:"toolchain_digest"`
	RunnerDigest      string `json:"runner_digest"`
	CPUCoreNS         int64  `json:"cpu_core_ns"`
	AllocatedCPUCount int64  `json:"allocated_cpu_count"`
	WallNS            int64  `json:"wall_ns"`
	PeakRSSBytes      int64  `json:"peak_rss_bytes"`
	ReadBytes         int64  `json:"read_bytes"`
	WriteBytes        int64  `json:"write_bytes"`
}
type Input struct {
	SchemaVersion            string            `json:"schema_version"`
	SnapshotDigest           string            `json:"snapshot_digest"`
	PolicyDigest             string            `json:"policy_digest"`
	RegistryDigest           string            `json:"registry_digest"`
	SelectionDigest          string            `json:"selection_digest"`
	ToolchainDigest          string            `json:"toolchain_digest"`
	RunnerDigest             string            `json:"runner_digest"`
	Obligations              []Obligation      `json:"obligations"`
	Commands                 []Command         `json:"commands"`
	ImpactedObligationIDs    []string          `json:"impacted_obligation_ids"`
	SelectedCommandIDs       []string          `json:"selected_command_ids"`
	SelectionReceipt         *SelectionReceipt `json:"selection_receipt"`
	FullOutcomes             []Outcome         `json:"full_outcomes"`
	SelectedOutcomes         []Outcome         `json:"selected_outcomes"`
	FullResourceReceipts     []ResourceReceipt `json:"full_resource_receipts"`
	SelectedResourceReceipts []ResourceReceipt `json:"selected_resource_receipts"`
	ExecutionAuthorized      bool              `json:"execution_authorized"`
	CIAuthorized             bool              `json:"ci_authorized"`
}
type Utilization struct {
	Numerator   int64 `json:"numerator"`
	Denominator int64 `json:"denominator"`
}
type ResourceTotals struct {
	CPUCoreNS    int64       `json:"cpu_core_ns"`
	PeakRSSBytes int64       `json:"peak_rss_bytes"`
	ReadBytes    int64       `json:"read_bytes"`
	WriteBytes   int64       `json:"write_bytes"`
	Utilization  Utilization `json:"utilization"`
}
type ResourceVector struct {
	Full     ResourceTotals `json:"full"`
	Selected ResourceTotals `json:"selected"`
	Class    ResourceClass  `json:"class"`
}
