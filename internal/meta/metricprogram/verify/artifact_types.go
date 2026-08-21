package verify

type programSelection struct {
	ProofChoice   string `json:"proof_choice"`
	Decision      string `json:"decision"`
	MetaOperation string `json:"meta_operation"`
	Reason        string `json:"reason"`
}

type coverage struct {
	BindingCount               int    `json:"binding_count"`
	ResolvedBindingCount       int    `json:"resolved_binding_count"`
	RegistryOperationCount     int    `json:"registry_operation_count"`
	ReferencedOperationCount   int    `json:"referenced_operation_count"`
	SelectionOperationResolved bool   `json:"selection_operation_resolved"`
	Status                     string `json:"status"`
}

type program struct {
	Schema                     string            `json:"schema"`
	Repository                 string            `json:"repository"`
	SubjectSHA                 string            `json:"subject_sha"`
	StrategyDigest             string            `json:"strategy_digest"`
	StrategyVerificationDigest string            `json:"strategy_verification_digest"`
	ExecutionPolicy            string            `json:"execution_policy"`
	RootPolicy                 rootPolicy        `json:"root_policy"`
	RegistryDigest             string            `json:"registry_digest"`
	SourcePath                 string            `json:"source_path"`
	SourceDigest               string            `json:"source_digest"`
	SemanticDigest             string            `json:"semantic_digest"`
	Operations                 []operationSpec   `json:"operations"`
	Bindings                   []resolvedBinding `json:"bindings"`
	Steps                      []programStep     `json:"steps"`
	Selection                  programSelection  `json:"selection"`
	Coverage                   coverage          `json:"coverage"`
	RepositoryWorkspaceWrites  bool              `json:"repository_workspace_writes"`
	PromotionAuthorized        bool              `json:"promotion_authorized"`
	Digest                     string            `json:"digest"`
}

type Report struct {
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
