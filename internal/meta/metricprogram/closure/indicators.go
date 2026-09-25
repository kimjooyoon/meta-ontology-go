package closure

func buildIndicators(receipt Receipt) ([]Indicator, error) {
	artifactDigest, err := digestValue(receipt.Artifact)
	if err != nil {
		return nil, err
	}
	return []Indicator{
		{
			ID: "MPC-FOUNDATION-ARTIFACT-001", Family: "FOUNDATION",
			Status: "SATISFIED", EvidenceDigest: artifactDigest,
		},
		{
			ID: "MPC-COHERENCE-META-CODE-002", Family: "COHERENCE",
			Status: "SATISFIED", EvidenceDigest: receipt.SourceDigest,
		},
		{
			ID: "MPC-COHERENCE-VERIFICATION-003", Family: "COHERENCE",
			Status: "SATISFIED", EvidenceDigest: receipt.VerificationDigest,
		},
		{
			ID: "MPC-REGRESSION-FIXED-POINT-004", Family: "REGRESSION",
			Status: "SATISFIED", EvidenceDigest: receipt.ProgramDigest,
		},
	}, nil
}
