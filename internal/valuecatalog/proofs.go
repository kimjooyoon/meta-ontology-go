package valuecatalog

func buildProofs(report Report) []Proof {
	extension := report.Improvement.After.Satisfied == 1
	return []Proof{
		{
			Choice: "FOUNDATION", Claim: "one source binds the baseline and an explicit extension slot",
			MetaOperation: "bind-catalog-baseline", EvidenceDigest: digestValue([]any{report.SourceDigest, report.OperationSpecs}),
			Passed: report.ActivitiesObserved == 2 && report.Improvement.Before == coordinate(0, 1) && report.OperationSpecMetrics.VerifiedTotal == OperationSpecAxisTotal,
		},
		{
			Choice: "COHERENCE", Claim: "the source-only program changes core IR and executes through the existing registry",
			MetaOperation: "compare-source-only-extension", EvidenceDigest: digestValue([]string{report.CoreIRFingerprint, report.Extension.SemanticFingerprint}),
			Passed: extension && report.ExtensionCoreProgram == "int.add:2" && report.Summary.CoreFingerprintSensitive == coordinate(1, 1) && report.OperationSpecMetrics.OpenClaims == 0,
		},
		{
			Choice: "REGRESSION", Claim: "baseline values and the unknown-operation counterexample remain exact",
			MetaOperation: "replay-catalog-regressions", EvidenceDigest: digestValue([]any{report.Baseline.Cases, report.Extension.Cases, report.Summary.UnknownCounterexamplePassed}),
			Passed: report.Baseline.Passed == 3 && report.Summary.UnknownCounterexamplePassed && (!extension || report.Extension.Passed == 3) && report.OperationSpecMetrics.UnknownPathCount == 0,
		},
	}
}
