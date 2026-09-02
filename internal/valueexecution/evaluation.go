package valueexecution

func evaluateProgram(report Report, source []byte, program Program) Report {
	report.Resolution = ResolutionBidirValue
	report.Activity = program.Activity
	report.ValueProgram = program.Text
	report.ValueProgramDigest = digestBytes([]byte(program.Text))
	report.SemanticFingerprint = program.SemanticFingerprint
	report.Registry = RegistrySummary{
		RegisteredOperations: len(operationRegistry), InvokedOperations: 1, OperationIDs: operationIDs(),
	}
	report.Cases = executeCases(program)
	report.Counterexamples = executeCounterexamples(report.SourcePath)
	measured := measure(report.SourcePath, source, program, report.Cases, report.Counterexamples)
	report.CoreIRFingerprint = measured.coreIRFingerprint
	report.Improvement = Improvement{
		ID: "value-level-computation", Before: coordinate(boolInt(measured.baselineReason == ""), 1),
		After:          coordinate(boolInt(measured.passedCases == len(report.Cases)), 1),
		BeforeEvidence: measured.baselineReason, AfterEvidence: digestValue(report.Cases),
	}
	report.Summary = Summary{
		ValueCasesPassed: measured.passedCases, ValueCasesTotal: len(report.Cases),
		CounterexamplesPassed: measured.passedCounterexamples, CounterexamplesTotal: len(report.Counterexamples),
		ValueOutputsObserved: measured.passedCases, DeterministicReplays: measured.replayedCases,
		RepositoryWrites:                 0,
		CoreIRProgramPreserved:           coordinate(boolInt(measured.coreIRProgramPreserved), 1),
		CoreIRFingerprintSensitive:       coordinate(boolInt(measured.coreIRFingerprintSensitive), 1),
		CoreIRUnknownAttributeFailClosed: coordinate(boolInt(measured.coreIRUnknownAttributeFailClosed), 1),
	}
	checks := measured.evidence(program, report.Counterexamples)
	report.Indicators = buildIndicators(checks)
	report.Views = buildViews(report.Indicators)
	report.Proofs = buildProofs(report, checks)
	if measured.coreIRProgramPreserved && measured.coreIRFingerprintSensitive && measured.coreIRUnknownAttributeFailClosed {
		report.Resolution = ResolutionCoreValue
	}
	if !allIndicatorsSatisfied(report.Indicators) {
		report.Reason = ReasonIndicatorUnsatisfied
		return finalize(report)
	}
	if measured.baselineReason != ReasonProgramMissing {
		report.Reason = measured.baselineReason
		return finalize(report)
	}
	report.Decision = DecisionProven
	report.Reason = ReasonExactWitness
	return finalize(report)
}
