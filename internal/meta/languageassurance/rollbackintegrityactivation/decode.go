package rollbackintegrityactivation

type assuranceReport struct {
	Schema              string                `json:"schema"`
	SubjectSHA          string                `json:"subject_sha"`
	DenominatorID       string                `json:"denominator_id"`
	AssuranceDecision   string                `json:"assurance_decision"`
	CandidateDecision   string                `json:"candidate_decision"`
	CandidateResolution string                `json:"candidate_resolution"`
	Summary             assuranceSummary      `json:"summary"`
	Obligations         []assuranceObligation `json:"obligations"`
	ReportDigest        string                `json:"report_digest"`
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

type eligibilityIndicator struct {
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Resolution    string `json:"resolution"`
	Satisfied     bool   `json:"satisfied"`
}

type eligibilityOperation struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}
