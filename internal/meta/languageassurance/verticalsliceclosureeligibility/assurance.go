package verticalsliceclosureeligibility

import "encoding/json"

type assuranceCapsule struct {
	Schema            string `json:"schema"`
	SubjectSHA        string `json:"subject_sha"`
	DenominatorID     string `json:"denominator_id"`
	DenominatorDigest string `json:"denominator_digest"`
	ReportDigest      string `json:"report_digest"`
	Summary           struct {
		DenominatorTotal          int `json:"denominator_total"`
		Operating                 int `json:"operating"`
		NotImplemented            int `json:"not_implemented"`
		ImplementationCoverageBPS int `json:"implementation_coverage_bps"`
		UnknownTopDecisions       int `json:"unknown_top_decisions"`
		RepositoryWrites          int `json:"repository_writes"`
	} `json:"summary"`
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
