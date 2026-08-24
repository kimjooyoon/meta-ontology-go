package externalcapabilityexecution

type Reference struct {
	RepositoryURL string `json:"repository_url"`
	CommitSHA     string `json:"commit_sha"`
	TreeSHA       string `json:"tree_sha"`
	GoVersion     string `json:"go_version"`
}

type ParentReport struct {
	Decision              string `json:"decision"`
	Resolution            string `json:"resolution"`
	Completed             int    `json:"completed"`
	Total                 int    `json:"total"`
	BasisPoints           int    `json:"basis_points"`
	OfficialMutationCount int    `json:"official_mutation_count"`
	PromotionCount        int    `json:"promotion_count"`
}

type CapabilityRun struct {
	RunID                 string `json:"run_id"`
	Status                string `json:"status"`
	Arithmetic            string `json:"arithmetic"`
	Function              string `json:"function"`
	EvaluatorExitCode     int    `json:"evaluator_exit_code"`
	MacroExitCode         int    `json:"macro_exit_code"`
	MacroGeneratedSHA256  string `json:"macro_generated_sha256"`
	MacroExpectedSHA256   string `json:"macro_expected_sha256"`
	NormalizedSHA256      string `json:"normalized_sha256"`
}
