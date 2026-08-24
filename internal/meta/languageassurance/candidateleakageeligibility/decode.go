package candidateleakageeligibility

import "encoding/json"

type assuranceCapsule struct {
	Schema, SubjectSHA, DenominatorID, DenominatorDigest string
	Obligations []struct {
		MetricID, Status, Resolution string
	} `json:"obligations"`
	Summary struct {
		DenominatorTotal, Operating, NotImplemented, ImplementationCoverageBPS int
	} `json:"summary"`
}

type shadowCapsule struct {
	Schema, SubjectSHA, DenominatorID, DenominatorDigest string
	Decision, Resolution string
	Summary struct {
		CasesTotal, CasesPassed, ExactPass, ExactFailClosed int
		InvariantFailClosed, CoverageBPS int
	} `json:"summary"`
	RepositoryWrites, PromotionCreditBPS int
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
