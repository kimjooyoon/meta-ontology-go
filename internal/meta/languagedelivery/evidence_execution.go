package languagedelivery

type ExecutionReceipt struct {
	Schema            string           `json:"schema"`
	HeadSHA           string           `json:"head_sha"`
	Decision          string           `json:"decision"`
	Resolution        string           `json:"resolution"`
	RepositoryWrites  int              `json:"repository_writes"`
	MutationAuthority bool             `json:"mutation_authority"`
	Summary           ExecutionSummary `json:"summary"`
}

type ExecutionSummary struct {
	CasesSatisfied       int `json:"cases_satisfied"`
	CasesTotal           int `json:"cases_total"`
	SourceExecutions     int `json:"source_executions"`
	DeterministicReplays int `json:"deterministic_replays"`
	DiagnosticRejections int `json:"diagnostic_rejections"`
	RepositoryWrites     int `json:"repository_writes"`
	MutationAuthorities  int `json:"mutation_authorities"`
}

func inspectExecution(data []byte, head string, receipt *ExecutionReceipt, entry ManifestEntry) SourceObservation {
	if err := unmarshalReceipt(data, receipt); err != nil {
		return unknownObservation(SourceExecution, entry, "SOURCE_JSON_UNKNOWN")
	}
	observation := baseObservation(SourceExecution, entry, receipt.Schema, receipt.Decision, receipt.Resolution)
	observation.RepositoryWrites = receipt.RepositoryWrites + receipt.Summary.RepositoryWrites
	observation.MutationAuthority = receipt.MutationAuthority || receipt.Summary.MutationAuthorities != 0
	if receipt.HeadSHA != head {
		return headUnknown(observation)
	}
	return finalizeObservation(observation, receipt.Schema, "gooo/language-source-execution-artifact/v1")
}
