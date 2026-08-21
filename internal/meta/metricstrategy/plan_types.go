package metricstrategy

type Candidate struct {
	ProofChoice      string   `json:"proof_choice"`
	IndicatorIDs     []string `json:"indicator_ids"`
	MetaOperations   []string `json:"meta_operations"`
	IndicatorCount   int      `json:"indicator_count"`
	UnsatisfiedCount int      `json:"unsatisfied_count"`
	Admissible       bool     `json:"admissible"`
	EvidenceDigest   string   `json:"evidence_digest"`
}

type StrategyPolicy struct {
	Schema         string   `json:"schema"`
	Choices        []string `json:"choices"`
	FailureRule    string   `json:"failure_rule"`
	FixedPointRule string   `json:"fixed_point_rule"`
}

type Selection struct {
	ProofChoice          string   `json:"proof_choice"`
	Decision             string   `json:"decision"`
	MetaOperation        string   `json:"meta_operation"`
	Reason               string   `json:"reason"`
	CandidateDigest      string   `json:"candidate_digest"`
	SourceMetaOperations []string `json:"source_meta_operations"`
}

type Plan struct {
	Schema                    string         `json:"schema"`
	Repository                string         `json:"repository"`
	SubjectSHA                string         `json:"subject_sha"`
	ExecutionPolicy           string         `json:"execution_policy"`
	Input                     InputEvidence  `json:"input"`
	RootPolicy                RootPolicy     `json:"root_policy"`
	Policy                    StrategyPolicy `json:"policy"`
	Bindings                  []Binding      `json:"bindings"`
	Candidates                []Candidate    `json:"candidates"`
	Selection                 Selection      `json:"selection"`
	RepositoryWorkspaceWrites bool           `json:"repository_workspace_writes"`
	PromotionAuthorized       bool           `json:"promotion_authorized"`
	Digest                    string         `json:"digest"`
}
