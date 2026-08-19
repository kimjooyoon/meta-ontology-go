package selectiveci

type wireCommandReceipt struct {
	CommandID             string `json:"command_id"`
	ReceiptID             string `json:"receipt_id"`
	Status                string `json:"status"`
	ProviderReceiptDigest string `json:"provider_receipt_digest"`
	PhaseReceiptDigest    string `json:"phase_receipt_digest"`
	ResourceReceiptDigest string `json:"resource_receipt_digest"`
	RegistryDigest        string `json:"registry_digest"`
	PlanDigest            string `json:"plan_digest"`
	Digest                string `json:"digest"`
}
type wireInput struct {
	Schema             string               `json:"schema"`
	Snapshots          wireSnapshotBinding  `json:"snapshots"`
	RegistryDigest     string               `json:"registry_digest"`
	PlanDigest         string               `json:"plan_digest"`
	ChangedRootIDs     []string             `json:"changed_root_ids"`
	SelectedCommandIDs []string             `json:"selected_command_ids"`
	ObligationIDs      []string             `json:"obligation_ids"`
	Paths              []wirePath           `json:"paths"`
	CommandReceipts    []wireCommandReceipt `json:"command_receipts"`
	EvidenceIDs        []string             `json:"evidence_ids"`
	InferencePath      wireInferencePath    `json:"inference_path"`
}
type wireSnapshotBinding struct {
	Base wireSnapshot `json:"base"`
	Head wireSnapshot `json:"head"`
}
type wireReceipt struct {
	Schema                  string              `json:"schema"`
	Status                  string              `json:"status"`
	Fallback                string              `json:"fallback"`
	Code                    string              `json:"code"`
	Snapshots               wireSnapshotBinding `json:"snapshots"`
	RegistryDigest          string              `json:"registry_digest"`
	PlanDigest              string              `json:"plan_digest"`
	SelectedCommandIDs      []string            `json:"selected_command_ids"`
	ObligationIDs           []string            `json:"obligation_ids"`
	PathIDs                 []string            `json:"path_ids"`
	RequiredCommandCount    int                 `json:"required_command_count"`
	RequiredObligationCount int                 `json:"required_obligation_count"`
	VerifiedCommandCount    int                 `json:"verified_command_count"`
	VerifiedObligationCount int                 `json:"verified_obligation_count"`
	VerifiedPathCount       int                 `json:"verified_path_count"`
	VerifiedCommandIDs      []string            `json:"verified_command_ids"`
	VerifiedObligationIDs   []string            `json:"verified_obligation_ids"`
	VerifiedPathIDs         []string            `json:"verified_path_ids"`
	Digest                  string              `json:"digest"`
}
