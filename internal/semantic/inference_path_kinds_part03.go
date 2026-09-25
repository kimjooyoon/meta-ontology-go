package semantic

func inferenceContractFor(kind InferenceKind) inferenceContract {
	switch kind {
	case InferenceAuthoritativeDeclaration:
		return inferenceContract{PhaseDeclaration, AuthoritySource, AuthorityDeclare, false, false, false}
	case InferenceDeterministicDerivation:
		return inferenceContract{PhaseDerivation, AuthoritySemantic, AuthorityDerive, false, false, false}
	case InferenceDerivedProjection:
		return inferenceContract{PhaseProjection, AuthorityDerived, AuthorityProject, false, false, true}
	case InferenceObservationCandidate:
		return inferenceContract{PhaseObservation, AuthorityCandidate, AuthorityObserve, true, false, false}
	case InferenceAcceptedLift:
		return inferenceContract{PhaseLift, AuthoritySemantic, AuthorityLift, false, true, false}
	default:
		return inferenceContract{PhaseVerification, AuthorityVerification, AuthorityVerify, false, true, false}
	}
}
