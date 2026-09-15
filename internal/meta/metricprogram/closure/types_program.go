package closure

type programDocument struct {
	Schema                     string              `json:"schema"`
	Repository                 string              `json:"repository"`
	SubjectSHA                 string              `json:"subject_sha"`
	StrategyDigest             string              `json:"strategy_digest"`
	StrategyVerificationDigest string              `json:"strategy_verification_digest"`
	ExecutionPolicy            string              `json:"execution_policy"`
	RootPolicy                 RootPolicy          `json:"root_policy"`
	RegistryDigest             string              `json:"registry_digest"`
	SourceDigest               string              `json:"source_digest"`
	SemanticDigest             string              `json:"semantic_digest"`
	Operations                 []operationDocument `json:"operations"`
	Bindings                   []bindingDocument   `json:"bindings"`
	Steps                      []stepDocument      `json:"steps"`
	Selection                  selectionDocument   `json:"selection"`
	Coverage                   coverageDocument    `json:"coverage"`
	RepositoryWorkspaceWrites  bool                `json:"repository_workspace_writes"`
	PromotionAuthorized        bool                `json:"promotion_authorized"`
	Digest                     string              `json:"digest"`
}
