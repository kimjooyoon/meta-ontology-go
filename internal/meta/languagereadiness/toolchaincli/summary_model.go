package toolchaincli

type Summary struct {
	Total               int `json:"total"`
	Executed            int `json:"executed"`
	Satisfied           int `json:"satisfied"`
	NotSatisfied        int `json:"not_satisfied"`
	Unresolved          int `json:"unresolved"`
	ReadinessBPS        int `json:"readiness_bps"`
	PositivePaths       int `json:"positive_paths"`
	GuardrailRejections int `json:"guardrail_rejections"`
	Invocations         int `json:"invocations"`
	DeclaredCommands    int `json:"declared_commands"`
	StructuredOutputs   int `json:"structured_outputs"`
	LanguageOperations  int `json:"language_operations"`
	ReplayMatches       int `json:"replay_matches"`
	BinaryBindings      int `json:"binary_bindings"`
	ExitMismatches      int `json:"exit_mismatches"`
	StdoutMismatches    int `json:"stdout_mismatches"`
	StderrMismatches    int `json:"stderr_mismatches"`
	ReplayMismatches    int `json:"replay_mismatches"`
	RepositoryWrites    int `json:"repository_writes"`
	RegistryDrift       int `json:"registry_drift"`
}
