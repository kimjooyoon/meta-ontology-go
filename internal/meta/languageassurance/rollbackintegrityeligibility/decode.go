package rollbackintegrityeligibility

import "encoding/json"

type assuranceCapsule struct {
	Schema              string                `json:"schema"`
	SubjectSHA          string                `json:"subject_sha"`
	DenominatorDigest   string                `json:"denominator_digest"`
	AssuranceDecision   string                `json:"assurance_decision"`
	CandidateDecision   string                `json:"candidate_decision"`
	CandidateResolution string                `json:"candidate_resolution"`
	Obligations         []assuranceObligation `json:"obligations"`
	Summary             assuranceSummary      `json:"summary"`
}

type assuranceObligation struct {
	MetricID      string  `json:"metric_id"`
	Status        string  `json:"status"`
	Resolution    string  `json:"resolution"`
	MetaOperation *string `json:"meta_operation"`
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

func decode[T any](payload []byte) (T, error) {
	var result T
	err := json.Unmarshal(payload, &result)
	return result, err
}
