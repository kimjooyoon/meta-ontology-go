package verify

const (
	strategySchema             = "gooo/metric-meta-strategy/v1"
	strategyVerificationSchema = "gooo/metric-meta-strategy-verification/v1"
	programSchema              = "gooo/metric-meta-program/v1"
	reportSchema               = "gooo/metric-meta-program-verification/v1"
	programSourceFilename      = "program.gooo"
)

type rootPolicy struct {
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	ReadmeRequirement     string `json:"readme_requirement"`
}

type strategyBinding struct {
	IndicatorID    string `json:"indicator_id"`
	Family         string `json:"family"`
	Trilemma       string `json:"trilemma"`
	MetaOperation  string `json:"meta_operation"`
	Expected       string `json:"expected"`
	Actual         string `json:"actual"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

type strategyCandidate struct {
	ProofChoice      string   `json:"proof_choice"`
	IndicatorIDs     []string `json:"indicator_ids"`
	MetaOperations   []string `json:"meta_operations"`
	IndicatorCount   int      `json:"indicator_count"`
	UnsatisfiedCount int      `json:"unsatisfied_count"`
	Admissible       bool     `json:"admissible"`
	EvidenceDigest   string   `json:"evidence_digest"`
}

type strategySelection struct {
	ProofChoice          string   `json:"proof_choice"`
	Decision             string   `json:"decision"`
	MetaOperation        string   `json:"meta_operation"`
	Reason               string   `json:"reason"`
	CandidateDigest      string   `json:"candidate_digest"`
	SourceMetaOperations []string `json:"source_meta_operations"`
}

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
