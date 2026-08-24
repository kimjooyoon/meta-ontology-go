package candidateleakageeligibility

import "encoding/json"

type assuranceCapsule struct {
	Schema            string `json:"schema"`
	SubjectSHA        string `json:"subject_sha"`
	DenominatorID     string `json:"denominator_id"`
	DenominatorDigest string `json:"denominator_digest"`
	Obligations       []struct {
		MetricID   string `json:"metric_id"`
		Status     string `json:"status"`
		Resolution string `json:"resolution"`
	} `json:"obligations"`
	Summary struct {
		DenominatorTotal          int `json:"denominator_total"`
		Operating                 int `json:"operating"`
		NotImplemented            int `json:"not_implemented"`
		ImplementationCoverageBPS int `json:"implementation_coverage_bps"`
	} `json:"summary"`
}

type shadowCapsule struct {
	Schema            string `json:"schema"`
	SubjectSHA        string `json:"subject_sha"`
	DenominatorID     string `json:"denominator_id"`
	DenominatorDigest string `json:"denominator_digest"`
	Decision          string `json:"decision"`
	Resolution        string `json:"resolution"`
	Summary           struct {
		CasesTotal          int `json:"cases_total"`
		CasesPassed         int `json:"cases_passed"`
		ExactPass           int `json:"exact_pass"`
		ExactFailClosed     int `json:"exact_fail_closed"`
		InvariantFailClosed int `json:"invariant_fail_closed"`
		CoverageBPS         int `json:"coverage_bps"`
	} `json:"summary"`
	RepositoryWrites   int `json:"repository_writes"`
	PromotionCreditBPS int `json:"promotion_credit_bps"`
}

func decodeAssurance(payload []byte) (assuranceCapsule, error) {
	var result assuranceCapsule
	err := json.Unmarshal(payload, &result)
	return result, err
}

func decodeShadow(payload []byte) (shadowCapsule, error) {
	var result shadowCapsule
	err := json.Unmarshal(payload, &result)
	return result, err
}

func candidateBaseline(capsule assuranceCapsule) bool {
	for _, obligation := range capsule.Obligations {
		if obligation.MetricID == MetricID {
			return obligation.Status == "NOT_IMPLEMENTED" && obligation.Resolution == "NONE"
		}
	}
	return false
}
