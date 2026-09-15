package adapter

func validateObserverMutation(observation NoWriteObservation) error {
	switch observation.Mutation.Status {
	case MutationEvidenceMissing:
		return oracleError(OracleNW001, "observer mutation-attempt evidence is missing")
	case MutationEvidenceUnverified:
		return oracleError(OracleNW003, "observer mutation-attempt evidence is not independently verified")
	case MutationEvidenceVerified:
		if observation.Mutation.Binding != observation.Binding {
			return oracleError(OracleID001, "mutation evidence binding does not match observer")
		}
		if err := validateVerifiedMutation(observation.Mutation, observation.Paths); err != nil {
			return oracleError(OracleNW003, "mutation evidence: "+err.Error())
		}
		if len(observation.Mutation.Attempts) != 0 {
			return oracleError(OracleNW004, "observer recorded an attempted mutation")
		}
		return nil
	default:
		return oracleError(OracleNW003, "observer mutation evidence status is invalid")
	}
}
