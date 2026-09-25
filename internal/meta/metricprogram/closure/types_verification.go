package closure

type verificationDocument struct {
	Schema                    string `json:"schema"`
	SubjectSHA                string `json:"subject_sha"`
	StrategyDigest            string `json:"strategy_digest"`
	ProgramDigest             string `json:"program_digest"`
	RegistryDigest            string `json:"registry_digest"`
	SourceDigest              string `json:"source_digest"`
	SemanticDigest            string `json:"semantic_digest"`
	BindingCount              int    `json:"binding_count"`
	OperationCount            int    `json:"operation_count"`
	StepCount                 int    `json:"step_count"`
	Status                    string `json:"status"`
	RepositoryWorkspaceWrites bool   `json:"repository_workspace_writes"`
	PromotionAuthorized       bool   `json:"promotion_authorized"`
	Digest                    string `json:"digest"`
}
