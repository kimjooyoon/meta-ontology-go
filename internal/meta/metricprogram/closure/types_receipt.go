package closure

type Receipt struct {
	Schema                    string           `json:"schema"`
	Repository                string           `json:"repository"`
	SubjectSHA                string           `json:"subject_sha"`
	RunID                     int64            `json:"run_id"`
	RunAttempt                int              `json:"run_attempt"`
	ExecutionPolicy           string           `json:"execution_policy"`
	RootPolicy                RootPolicy       `json:"root_policy"`
	Artifact                  ArtifactIdentity `json:"artifact"`
	Files                     Files            `json:"files"`
	ProgramDigest             string           `json:"program_digest"`
	VerificationDigest        string           `json:"verification_digest"`
	StrategyDigest            string           `json:"strategy_digest"`
	RegistryDigest            string           `json:"registry_digest"`
	SourceDigest              string           `json:"source_digest"`
	SemanticDigest            string           `json:"semantic_digest"`
	Indicators                []Indicator      `json:"indicators"`
	BindingCount              int              `json:"binding_count"`
	OperationCount            int              `json:"operation_count"`
	StepCount                 int              `json:"step_count"`
	Status                    string           `json:"status"`
	WriteEffect               string           `json:"write_effect"`
	RepositoryWorkspaceWrites bool             `json:"repository_workspace_writes"`
	PromotionAuthorized       bool             `json:"promotion_authorized"`
	Digest                    string           `json:"digest"`
}
