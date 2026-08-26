package verify

func expectedIndicators(value receipt) ([]indicator, error) {
	artifactDigest, err := digestValue(value.Artifact)
	if err != nil {
		return nil, err
	}
	return []indicator{
		{
			ID: "MPC-FOUNDATION-ARTIFACT-001", Family: "FOUNDATION",
			Status: "SATISFIED", EvidenceDigest: artifactDigest,
		},
		{
			ID: "MPC-COHERENCE-META-CODE-002", Family: "COHERENCE",
			Status: "SATISFIED", EvidenceDigest: value.SourceDigest,
		},
		{
			ID: "MPC-COHERENCE-VERIFICATION-003", Family: "COHERENCE",
			Status: "SATISFIED", EvidenceDigest: value.VerificationDigest,
		},
		{
			ID: "MPC-REGRESSION-FIXED-POINT-004", Family: "REGRESSION",
			Status: "SATISFIED", EvidenceDigest: value.ProgramDigest,
		},
	}, nil
}
