package languagepackageruntime

func proofs(report Report) []Proof {
	foundation := report.Source.ObservationKnown && report.Summary.Executed == 18 &&
		report.Summary.Packages == 40 && report.Summary.Sources == 50 && report.Summary.SemanticBindings == 50
	coherence := report.Summary.PositivePaths == 10 && report.Summary.Imports == 40 &&
		report.Summary.Initializations == 40 && report.Summary.EntryBindings == 10 &&
		report.Summary.CanonicalReplays == 10 && report.Summary.OrderInvariantReplays == 3
	regression := report.Summary.GuardrailRejections == 8 && report.Summary.UnknownObservations == 0 &&
		report.Summary.InvalidAcceptances == 0 && report.Summary.EffectfulOperations == 0 &&
		report.RepositoryWrites == 0 && !report.MutationAuthorized
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-versioned-package-runtime", EvidenceDigest: digestValue(report.Source), Passed: foundation},
		{Choice: "COHERENCE", MetaOperation: "replay-package-runtime-image", EvidenceDigest: digestValue(report.Cases), Passed: coherence},
		{Choice: "REGRESSION", MetaOperation: "reject-invalid-package-runtime", EvidenceDigest: digestValue(report.Summary), Passed: regression},
	}
}
