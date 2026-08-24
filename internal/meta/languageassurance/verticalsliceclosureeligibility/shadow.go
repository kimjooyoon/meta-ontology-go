package verticalsliceclosureeligibility

import "encoding/json"

type assuranceCapsule struct {
	Schema string `json:"schema"`
	SubjectSHA string `json:"subject_sha"`
	DenominatorID string `json:"denominator_id"`
	DenominatorDigest string `json:"denominator_digest"`
	ReportDigest string `json:"report_digest"`
	Summary struct {
		DenominatorTotal int `json:"denominator_total"`
		Operating int `json:"operating"`
		NotImplemented int `json:"not_implemented"`
		ImplementationCoverageBPS int `json:"implementation_coverage_bps"`
		UnknownTopDecisions int `json:"unknown_top_decisions"`
		RepositoryWrites int `json:"repository_writes"`
	} `json:"summary"`
}

type boundary struct {
	ID string `json:"id"`
	Status string `json:"status"`
	Resolution string `json:"resolution"`
	Value int `json:"value"`
	Target int `json:"target"`
	LinksSatisfied int `json:"links_satisfied"`
	LinksTotal int `json:"links_total"`
	HeadSHA string `json:"head_sha"`
	EvidenceAvailable bool `json:"evidence_available"`
	UnknownTopDecision bool `json:"unknown_top_decision"`
	KnownFailure bool `json:"known_failure"`
	RepositoryWrites int `json:"repository_writes"`
}

type sourceIndicator struct {
	Class string `json:"class"`
	ProofChoice string `json:"proof_choice"`
	Satisfied bool `json:"satisfied"`
}

type shadowCapsule struct {
	Schema string `json:"schema"`
	MetricID string `json:"metric_id"`
	MetaOperation string `json:"meta_operation"`
	Decision string `json:"decision"`
	Reason string `json:"reason"`
	Resolution string `json:"resolution"`
	EnforcementEffect string `json:"enforcement_effect"`
	HeadSHA string `json:"head_sha"`
	AssuranceSubjectSHA string `json:"assurance_subject_sha"`
	AssuranceDigest string `json:"assurance_digest"`
	DenominatorDigest string `json:"denominator_digest"`
	Summary struct {
		DenominatorTotal int `json:"denominator_total"`
		BeforeOperating int `json:"before_operating"`
		ProjectedOperating int `json:"projected_operating"`
		BeforeCoverageBPS int `json:"before_coverage_bps"`
		ProjectedCoverageBPS int `json:"projected_coverage_bps"`
		BoundariesTotal int `json:"boundaries_total"`
		BoundariesSatisfied int `json:"boundaries_satisfied"`
		UnknownBoundaries int `json:"unknown_boundaries"`
		BlockedBoundaries int `json:"blocked_boundaries"`
		LinksTotal int `json:"links_total"`
		LinksSatisfied int `json:"links_satisfied"`
		EvidenceAvailable int `json:"evidence_available"`
		UnknownTopDecisions int `json:"unknown_top_decisions"`
		KnownFailures int `json:"known_failures"`
		ObservedRepositoryWrites int `json:"observed_repository_writes"`
	} `json:"summary"`
	Boundaries []boundary `json:"boundaries"`
	Indicators []sourceIndicator `json:"indicators"`
	RepositoryWrites int `json:"repository_writes"`
	PromotionApplied int `json:"promotion_applied"`
	ReportDigest string `json:"report_digest"`
}

func decodeEvidence(input Input) (assuranceCapsule, shadowCapsule, error) {
	var assurance assuranceCapsule
	var shadow shadowCapsule
	if err := json.Unmarshal(input.Assurance.Payload, &assurance); err != nil {
		return assurance, shadow, err
	}
	err := json.Unmarshal(input.Shadow.Payload, &shadow)
	return assurance, shadow, err
}
