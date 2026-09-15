package candidateleakageactivation

type assuranceReport struct {
	Schema      string                `json:"schema"`
	SubjectSHA  string                `json:"subject_sha"`
	Summary     assuranceSummary      `json:"summary"`
	Obligations []assuranceObligation `json:"obligations"`
}

type assuranceSummary struct {
	DenominatorTotal          int `json:"denominator_total"`
	Operating                 int `json:"operating"`
	NotImplemented            int `json:"not_implemented"`
	ImplementationCoverageBPS int `json:"implementation_coverage_bps"`
	UnresolvedIndicators      int `json:"unresolved_indicators"`
	ViolatedGuardrails        int `json:"violated_guardrails"`
	RepositoryWrites          int `json:"repository_writes"`
}

type assuranceObligation struct {
	MetricID      string `json:"metric_id"`
	Status        string `json:"status"`
	Resolution    string `json:"resolution"`
	MetaOperation string `json:"meta_operation"`
}
