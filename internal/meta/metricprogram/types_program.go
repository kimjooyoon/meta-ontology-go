package metricprogram

type OperationSpec struct {
	ID                  string `json:"id"`
	Activity            string `json:"activity"`
	ProofChoice         string `json:"proof_choice"`
	InputEntity         string `json:"input_entity"`
	OutputEntity        string `json:"output_entity"`
	Mode                string `json:"mode"`
	Ordinal             int    `json:"ordinal"`
	RepositoryWrites    bool   `json:"repository_writes"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
}

type ResolvedBinding struct {
	IndicatorID     string `json:"indicator_id"`
	ProofChoice     string `json:"proof_choice"`
	OperationID     string `json:"operation_id"`
	Activity        string `json:"activity"`
	Mode            string `json:"mode"`
	EvidenceDigest  string `json:"evidence_digest"`
	OperationDigest string `json:"operation_digest"`
}

type ProgramStep struct {
	Index           int      `json:"index"`
	OperationID     string   `json:"operation_id"`
	Activity        string   `json:"activity"`
	Mode            string   `json:"mode"`
	DependsOn       []string `json:"depends_on"`
	InputEntity     string   `json:"input_entity"`
	OutputEntity    string   `json:"output_entity"`
	OperationDigest string   `json:"operation_digest"`
}

type ProgramSelection struct {
	ProofChoice   string `json:"proof_choice"`
	Decision      string `json:"decision"`
	MetaOperation string `json:"meta_operation"`
	Reason        string `json:"reason"`
}

type Coverage struct {
	BindingCount               int    `json:"binding_count"`
	ResolvedBindingCount       int    `json:"resolved_binding_count"`
	RegistryOperationCount     int    `json:"registry_operation_count"`
	ReferencedOperationCount   int    `json:"referenced_operation_count"`
	SelectionOperationResolved bool   `json:"selection_operation_resolved"`
	Status                     string `json:"status"`
}

type Program struct {
	Schema                     string            `json:"schema"`
	Repository                 string            `json:"repository"`
	SubjectSHA                 string            `json:"subject_sha"`
	StrategyDigest             string            `json:"strategy_digest"`
	StrategyVerificationDigest string            `json:"strategy_verification_digest"`
	ExecutionPolicy            string            `json:"execution_policy"`
	RootPolicy                 RootPolicy        `json:"root_policy"`
	RegistryDigest             string            `json:"registry_digest"`
	SourcePath                 string            `json:"source_path"`
	SourceDigest               string            `json:"source_digest"`
	SemanticDigest             string            `json:"semantic_digest"`
	Operations                 []OperationSpec   `json:"operations"`
	Bindings                   []ResolvedBinding `json:"bindings"`
	Steps                      []ProgramStep     `json:"steps"`
	Selection                  ProgramSelection  `json:"selection"`
	Coverage                   Coverage          `json:"coverage"`
	RepositoryWorkspaceWrites  bool              `json:"repository_workspace_writes"`
	PromotionAuthorized        bool              `json:"promotion_authorized"`
	Digest                     string            `json:"digest"`
}
