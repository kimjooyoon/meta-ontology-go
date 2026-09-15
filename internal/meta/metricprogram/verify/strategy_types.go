package verify

type strategyInput struct {
	SourceIndicatorSchema string `json:"source_indicator_schema"`
	SourcePolicySchema    string `json:"source_policy_schema"`
	SourceMetricsDigest   string `json:"source_metrics_digest"`
	InterventionSchema    string `json:"intervention_schema"`
	InterventionDigest    string `json:"intervention_digest"`
	VerificationSchema    string `json:"verification_schema"`
	VerificationDigest    string `json:"verification_digest"`
	IndicatorCount        int    `json:"indicator_count"`
	ProjectionCount       int    `json:"projection_count"`
}

type strategyPolicy struct {
	Schema         string   `json:"schema"`
	Choices        []string `json:"choices"`
	FailureRule    string   `json:"failure_rule"`
	FixedPointRule string   `json:"fixed_point_rule"`
}

type strategyPlan struct {
	Schema                    string              `json:"schema"`
	Repository                string              `json:"repository"`
	SubjectSHA                string              `json:"subject_sha"`
	ExecutionPolicy           string              `json:"execution_policy"`
	Input                     strategyInput       `json:"input"`
	RootPolicy                rootPolicy          `json:"root_policy"`
	Policy                    strategyPolicy      `json:"policy"`
	Bindings                  []strategyBinding   `json:"bindings"`
	Candidates                []strategyCandidate `json:"candidates"`
	Selection                 strategySelection   `json:"selection"`
	RepositoryWorkspaceWrites bool                `json:"repository_workspace_writes"`
	PromotionAuthorized       bool                `json:"promotion_authorized"`
	Digest                    string              `json:"digest"`
}

type strategyVerification struct {
	Schema                    string `json:"schema"`
	PlanDigest                string `json:"plan_digest"`
	SourceMetricsDigest       string `json:"source_metrics_digest"`
	InterventionDigest        string `json:"intervention_digest"`
	BindingCount              int    `json:"binding_count"`
	CandidateCount            int    `json:"candidate_count"`
	SelectedProofChoice       string `json:"selected_proof_choice"`
	Status                    string `json:"status"`
	RepositoryWorkspaceWrites bool   `json:"repository_workspace_writes"`
	PromotionAuthorized       bool   `json:"promotion_authorized"`
	Digest                    string `json:"digest"`
}
