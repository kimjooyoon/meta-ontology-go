package languagedelivery

func inspectJourney(data []byte, head string, receipt *JourneyReceipt, entry ManifestEntry) SourceObservation {
	if err := unmarshalReceipt(data, receipt); err != nil {
		return unknownObservation(SourceUserJourney, entry, "SOURCE_JSON_UNKNOWN")
	}
	observation := baseObservation(SourceUserJourney, entry, receipt.Schema, receipt.Decision, receipt.Resolution)
	observation.RepositoryWrites = receipt.RepositoryWrites
	observation.MutationAuthority = receipt.MutationAuthority
	if receipt.Source.ExpectedHeadSHA != head {
		return headUnknown(observation)
	}
	return finalizeObservation(observation, receipt.Schema, "gooo/user-journey-scorecard/v1")
}

func inspectConformance(data []byte, head string, receipt *ConformanceReceipt, entry ManifestEntry) SourceObservation {
	if err := unmarshalReceipt(data, receipt); err != nil {
		return unknownObservation(SourceConformance, entry, "SOURCE_JSON_UNKNOWN")
	}
	observation := baseObservation(SourceConformance, entry, receipt.Schema, receipt.Decision, receipt.Resolution)
	observation.RepositoryWrites = receipt.RepositoryWrites + receipt.Summary.RepositoryWrites
	observation.MutationAuthority = receipt.MutationAuthorized || receipt.Summary.MutationAuthorities != 0
	if len(receipt.Surfaces) == 0 {
		return headUnknown(observation)
	}
	for _, surface := range receipt.Surfaces {
		if surface.HeadSHA != head {
			return headUnknown(observation)
		}
	}
	return finalizeObservation(observation, receipt.Schema, "gooo/toolchain-conformance-report/v1")
}
