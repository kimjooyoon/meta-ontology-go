package proposalcompat

type LegacySummary struct {
	Satisfied        int `json:"satisfied"`
	Total            int `json:"total"`
	Unresolved       int `json:"unresolved"`
	RepositoryWrites int `json:"repository_writes"`
}

type LegacyReceipt struct {
	Schema         string        `json:"schema"`
	CurrentHeadSHA string        `json:"current_head_sha"`
	Decision       string        `json:"decision"`
	Summary        LegacySummary `json:"summary"`
	ReportDigest   string        `json:"report_digest"`
}

type Source struct {
	ExpectedHeadSHA          string `json:"expected_head_sha"`
	SourceSchema             string `json:"source_schema"`
	SourceDecision           string `json:"source_decision"`
	SourceReportDigest       string `json:"source_report_digest"`
	SourceFileSHA256         string `json:"source_file_sha256"`
	SourceSatisfied          int    `json:"source_satisfied"`
	SourceTotal              int    `json:"source_total"`
	SourceUnresolved         int    `json:"source_unresolved"`
	SourceRepositoryWrites   int    `json:"source_repository_writes"`
	SourceMutationAuthorized bool   `json:"source_mutation_authorized"`
	TargetSchema             string `json:"target_schema"`
	TargetReportDigest       string `json:"target_report_digest"`
	TargetFileSHA256         string `json:"target_file_sha256"`
	ProjectedFields          int    `json:"projected_fields"`
	FieldLosses              int    `json:"field_losses"`
	RepositoryWrites         int    `json:"repository_writes"`
}
