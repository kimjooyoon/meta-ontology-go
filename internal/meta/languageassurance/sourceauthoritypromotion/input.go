package sourceauthoritypromotion

type assuranceDocument struct {
	Schema            string                `json:"schema"`
	SubjectSHA        string                `json:"subject_sha"`
	DenominatorID     string                `json:"denominator_id"`
	DenominatorDigest string                `json:"denominator_digest"`
	AssuranceDecision string                `json:"assurance_decision"`
	CandidateDecision string                `json:"candidate_decision"`
	Denominator       []assuranceDefinition `json:"denominator"`
	Obligations       []assuranceObligation `json:"obligations"`
	Summary           assuranceSummary      `json:"summary"`
}

type assuranceDefinition struct {
	MetricID              string `json:"metric_id"`
	Priority              string `json:"priority"`
	Class                 string `json:"class"`
	ProofChoice           string `json:"proof_choice"`
	RequiredMetaOperation string `json:"required_meta_operation"`
}

type assuranceObligation struct {
	MetricID      string `json:"metric_id"`
	Status        string `json:"status"`
	Resolution    string `json:"resolution"`
	MetaOperation string `json:"meta_operation"`
}

type assuranceSummary struct {
	DenominatorTotal          int `json:"denominator_total"`
	Operating                 int `json:"operating"`
	NotImplemented            int `json:"not_implemented"`
	ImplementationCoverageBPS int `json:"implementation_coverage_bps"`
	UnknownTopDecisions       int `json:"unknown_top_decisions"`
	UnresolvedIndicators      int `json:"unresolved_indicators"`
	ViolatedGuardrails        int `json:"violated_guardrails"`
	RepositoryWrites          int `json:"repository_writes"`
}
