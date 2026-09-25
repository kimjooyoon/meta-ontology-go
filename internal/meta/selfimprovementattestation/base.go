package selfimprovementattestation

func baseReceipt(request Request) ResolutionReceipt {
	prior := request.TransportReceipt
	return ResolutionReceipt{
		Schema:              resolutionSchema,
		Metaprogram:         metaprogram,
		MetricID:            metricID,
		Contract:            prior.Contract,
		SubjectSHA:          prior.SubjectSHA,
		PriorReceiptDigest:  prior.Digest,
		SourceArchiveDigest: request.ArchiveDigest,
		Coordinate:          Coordinate{Stage: "ATTEST", Step: "verify-producer-identity"},
		Checker: Checker{
			Name:                "gh attestation verify",
			Version:             request.VerifierVersion,
			ExitCode:            request.VerifierExitCode,
			VerifiedResultTotal: len(request.Verification),
		},
		Obligations:       append([]Obligation(nil), prior.Obligations...),
		OpenObligationIDs: append([]string(nil), prior.OpenObligationIDs...),
		PriorMetrics:      prior.Metrics,
		Metrics:           prior.Metrics,
		Authority:         Authority{},
		NotClaimed: []string{
			"attestation-proves-program-correctness",
			"predicate-fields-are-independent-of-the-producer",
			"whole-language-transport-complete",
		},
	}
}

func lowerResolutionViews() []ReaderView {
	return []ReaderView{
		{Audience: "NON_ATTESTING_READER", Resolution: "LOWER_RESOLUTION", VerifiedTotal: 7, FixedTotal: 8, CoverageBasisPoints: 8750},
		{Audience: "ATTESTATION_READER", Resolution: "LOWER_RESOLUTION", VerifiedTotal: 7, FixedTotal: 8, CoverageBasisPoints: 8750},
	}
}
