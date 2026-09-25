package verify

type Report struct {
	Schema                    string `json:"schema"`
	SubjectSHA                string `json:"subject_sha"`
	ReceiptDigest             string `json:"receipt_digest"`
	ArtifactID                int64  `json:"artifact_id"`
	ArtifactDigest            string `json:"artifact_digest"`
	FileCount                 int    `json:"file_count"`
	BindingCount              int    `json:"binding_count"`
	OperationCount            int    `json:"operation_count"`
	StepCount                 int    `json:"step_count"`
	Status                    string `json:"status"`
	RepositoryWorkspaceWrites bool   `json:"repository_workspace_writes"`
	PromotionAuthorized       bool   `json:"promotion_authorized"`
	Digest                    string `json:"digest"`
}
