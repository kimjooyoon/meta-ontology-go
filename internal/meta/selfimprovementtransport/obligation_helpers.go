package selfimprovementtransport

func knownObligation(id, route, stage, step string, passed bool,
	successReason, failureReason string, evidence any) Obligation {
	status, reason := StatusFalse, failureReason
	if passed {
		status, reason = StatusVerified, successReason
	}
	return Obligation{
		ID: id, ProofRoute: route, Coordinate: Coordinate{Stage: stage, Step: step},
		Status: status, Reason: reason, EvidenceDigest: digestJSON(evidence),
	}
}

func attestationObligationFor(attestation Attestation) Obligation {
	obligation := Obligation{
		ID: attestationObligation, ProofRoute: "COHERENCE",
		Coordinate: Coordinate{Stage: "ATTEST", Step: "verify-producer-identity"},
	}
	switch attestation.Status {
	case "VERIFIED":
		if validDigest(attestation.Digest) && attestation.ProducerIdentity != "" {
			obligation.Status, obligation.Reason = StatusVerified, "PRODUCER_ATTESTATION_VERIFIED"
			obligation.EvidenceDigest = digestJSON(attestation)
			return obligation
		}
		obligation.Status, obligation.Reason = StatusFalse, "PRODUCER_ATTESTATION_INVALID"
		obligation.EvidenceDigest = digestJSON(attestation)
	case "", "UNKNOWN":
		obligation.Status, obligation.Reason = StatusUnknown, ReasonAttestation
	default:
		obligation.Status, obligation.Reason = StatusFalse, "PRODUCER_ATTESTATION_REJECTED"
		obligation.EvidenceDigest = digestJSON(attestation)
	}
	return obligation
}
