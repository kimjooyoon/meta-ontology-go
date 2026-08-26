package valueexecution

func buildProofs(report Report, checks evidence) []Proof {
	return []Proof{
		{
			Choice: "FOUNDATION", Claim: "the explicit Gooo program survives core lowering while unknown attributes fail closed",
			MetaOperation: "bind-explicit-source-to-core-ir", EvidenceDigest: report.SourceDigest,
			Passed: checks.sourceParsed && checks.programPresent && checks.coreIRProgramPreserved && checks.coreIRUnknownAttributeFailClosed,
		},
		{
			Choice: "COHERENCE", Claim: "bidir and core fingerprints vary with the registry-bound program",
			MetaOperation: "compile-registry-bound-program",
			EvidenceDigest: digestValue([]string{report.SemanticFingerprint, report.CoreIRFingerprint, report.ValueProgramDigest}),
			Passed: checks.semanticBound && checks.fingerprintSensitive && checks.coreIRFingerprintSensitive &&
				checks.registryKnown && checks.operandParsed && checks.signatureSupported,
		},
		{
			Choice: "REGRESSION", Claim: "fixed value cases and fail-closed counterexamples replay exactly",
			MetaOperation: "replay-value-witness-corpus", EvidenceDigest: digestValue(struct {
				Cases           []CaseResult
				Counterexamples []CounterexampleResult
			}{report.Cases, report.Counterexamples}),
			Passed: checks.valueCasesExact && checks.outputsObserved && checks.deterministicReplay &&
				checks.counterexamplesExact && checks.overflowFailClosed,
		},
	}
}
