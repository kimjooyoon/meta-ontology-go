package languagesemanticbinding

type conceptArtifact struct {
	Schema           string        `json:"schema"`
	Decision         string        `json:"decision"`
	ReplayEqual      bool          `json:"replay_equal"`
	RepositoryWrites int           `json:"repository_writes"`
	ArtifactDigest   string        `json:"artifact_digest"`
	Report           conceptReport `json:"report"`
}

type conceptReport struct {
	Decision string              `json:"decision"`
	Summary  conceptSummary      `json:"summary"`
	Concepts []conceptDefinition `json:"concepts"`
}

type conceptSummary struct {
	Concepts                int `json:"concepts"`
	CodeBound               int `json:"code_bound"`
	UseCaseBound            int `json:"use_case_bound"`
	MetricBound             int `json:"metric_bound"`
	Operating               int `json:"operating"`
	Conformed               int `json:"conformed"`
	Unbound                 int `json:"unbound"`
	UnverifiedNoveltyClaims int `json:"unverified_novelty_claims"`
	RepositoryWrites        int `json:"repository_writes"`
}

type conceptDefinition struct {
	ID             string           `json:"id"`
	MetaOperation  string           `json:"meta_operation"`
	Stage          string           `json:"stage"`
	NoveltyClaim   bool             `json:"novelty_claim"`
	CodeBindings   []string         `json:"code_bindings"`
	MetricBindings []string         `json:"metric_bindings"`
	UseCases       []conceptUseCase `json:"use_cases"`
}

type conceptUseCase struct {
	ID              string `json:"id"`
	ExpectedOutcome string `json:"expected_outcome"`
}
