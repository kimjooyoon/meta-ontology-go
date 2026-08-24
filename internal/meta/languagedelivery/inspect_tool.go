package languagedelivery

func inspectLSP(data []byte, head string, receipt *LSPReceipt, entry ManifestEntry) SourceObservation {
	if err := unmarshalReceipt(data, receipt); err != nil {
		return unknownObservation(SourceLSP, entry, "SOURCE_JSON_UNKNOWN")
	}
	observation := baseObservation(SourceLSP, entry, receipt.Schema, receipt.Decision, receipt.Resolution)
	observation.RepositoryWrites = receipt.Summary.RepositoryWrites
	observation.MutationAuthority = receipt.Summary.MutationAuthorities != 0
	if receipt.HeadSHA != "" && receipt.HeadSHA != head {
		return headUnknown(observation)
	}
	return finalizeObservation(observation, receipt.Schema, "gooo/toolchain-lsp-report/v1")
}

func inspectRelease(data []byte, head string, receipt *ReleaseReceipt, entry ManifestEntry) SourceObservation {
	if err := unmarshalReceipt(data, receipt); err != nil {
		return unknownObservation(SourceRelease, entry, "SOURCE_JSON_UNKNOWN")
	}
	observation := baseObservation(SourceRelease, entry, receipt.Schema, receipt.Decision, receipt.Resolution)
	observation.RepositoryWrites = receipt.Summary.RepositoryWrites
	observation.MutationAuthority = receipt.Summary.MutationAuthorities != 0
	if receipt.HeadSHA != "" && receipt.HeadSHA != head {
		return headUnknown(observation)
	}
	return finalizeObservation(observation, receipt.Schema, "gooo/toolchain-cross-platform-release-report/v1")
}

func inspectReadiness(data []byte, head string, receipt *ReadinessArtifact, entry ManifestEntry) SourceObservation {
	if err := unmarshalReceipt(data, receipt); err != nil {
		return unknownObservation(SourceReadiness, entry, "SOURCE_JSON_UNKNOWN")
	}
	observation := baseObservation(SourceReadiness, entry, receipt.Schema, receipt.Decision, "EXACT")
	observation.RepositoryWrites = receipt.Report.RepositoryWrites
	if receipt.HeadSHA != "" && receipt.HeadSHA != head {
		return headUnknown(observation)
	}
	if receipt.Report.Decision != "PASS" {
		observation.Decision = receipt.Report.Decision
	}
	return finalizeObservation(observation, receipt.Schema, "gooo/language-readiness-artifact/v1")
}
