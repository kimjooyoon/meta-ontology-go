package repositorytopology

type SourceIndicator struct {
	MetricID            string `json:"metric_id"`
	Family              string `json:"family"`
	Subject             string `json:"subject"`
	SubjectKind         string `json:"subject_kind"`
	Value               int    `json:"value"`
	Limit               int    `json:"limit"`
	Relation            string `json:"relation"`
	Applicability       string `json:"applicability"`
	ApplicabilityRuleID string `json:"applicability_rule_id"`
	ApplicabilityReason string `json:"applicability_reason"`
	Blocking            bool   `json:"blocking"`
	Satisfied           bool   `json:"satisfied"`
	ProofChoice         string `json:"proof_choice"`
	Producer            string `json:"producer"`
	Consumer            string `json:"consumer"`
	MetaOperation       string `json:"meta_operation"`
	Detail              string `json:"detail"`
	Decision            string `json:"decision"`
	EvaluationState     string `json:"evaluation_state"`
	FailureReason       string `json:"failure_reason"`
	FailureCode         string `json:"failure_code"`
	EnforcementEffect   string `json:"enforcement_effect"`
}
