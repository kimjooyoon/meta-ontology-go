package semanticdeltareceipt

func Evaluate(input Input) Report {
	receipt, err := Produce(input)
	if err != nil {
		receipt = Receipt{Schema: ReceiptSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA,
			Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: "COHERENCE",
			Stage: "produce", Step: "error", Reason: ReasonSubject, Decision: DecisionFailClosed,
			Resolution: ResolutionUnknown, Classification: ClassIndeterminate, RepositoryWrites: 0}
		sealReceipt(&receipt)
	}
	verdict := Adjudicate(input, receipt)
	summary := Summary{CasesTotal: 1, CasesPassed: boolToInt(verdict.Passed), AdjudicatedCases: boolToInt(verdict.Passed), RepositoryWrites: receipt.RepositoryWrites}
	if receipt.TextualDelta.Changed {
		summary.TextualChanges = 1
	}
	if receipt.StructuralDelta.Status != "" {
		summary.StructuralObservations = 1
	}
	if len(receipt.ClaimTransitions) > 0 {
		summary.ClaimTransitionCases = 1
	}
	switch receipt.Classification {
	case ClassPreserved:
		summary.SemanticPreserved = 1
	case ClassChanged:
		summary.SemanticChanged = 1
	case ClassIndeterminate:
		summary.Indeterminate, summary.UnknownPaths = 1, 1
	}
	report := Report{Schema: ReportSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA,
		Receipt: receipt, IndependentVerdict: verdict, Indicators: operationBindings(summary), RepositoryWrites: 0}
	sealReport(&report)
	return report
}

func mismatchVerdict() Verdict {
	return Verdict{Decision: DecisionFailClosed, Resolution: ResolutionInvariant, Classification: ClassIndeterminate,
		Reason: ReasonReceipt, Passed: false, Producer: Producer, Consumer: Consumer,
		Stage: "adjudicate", Step: "reject-mismatch"}
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func RunSuite(subjectSHA string) Suite {
	definitions := Denominator()
	cases := make([]CaseResult, 0, len(definitions))
	summary := Summary{CasesTotal: len(definitions)}
	passed := 0
	for _, definition := range definitions {
		report := Evaluate(CaseInput(definition.ID, subjectSHA))
		casePassed := report.IndependentVerdict.Passed && report.IndependentVerdict.Decision == definition.ExpectedDecision &&
			report.IndependentVerdict.Resolution == definition.ExpectedResolution && report.IndependentVerdict.Classification == definition.ExpectedClass &&
			report.IndependentVerdict.Reason == definition.ExpectedReason
		if casePassed {
			passed++
		}
		cases = append(cases, CaseResult{Definition: definition, Passed: casePassed, Report: report})
		mergeSummary(&summary, report, casePassed)
	}
	decision, resolution := DecisionFailClosed, ResolutionInvariant
	if passed == len(definitions) {
		decision, resolution = DecisionFixedPoint, ResolutionExact
	}
	summary.CasesPassed = passed
	suite := Suite{Schema: SuiteSchema, SubjectSHA: subjectSHA, DenominatorID: DenominatorID,
		DenominatorDigest: digestValue(definitions), Decision: decision, Resolution: resolution,
		Cases: cases, Summary: summary, CoverageBPS: ratio(passed, len(definitions))}
	sealSuite(&suite)
	return suite
}

func mergeSummary(summary *Summary, report Report, passed bool) {
	if report.Receipt.TextualDelta.Changed {
		summary.TextualChanges++
	}
	if report.Receipt.StructuralDelta.Status != "" {
		summary.StructuralObservations++
	}
	if len(report.Receipt.ClaimTransitions) > 0 {
		summary.ClaimTransitionCases++
	}
	if report.IndependentVerdict.Passed {
		summary.AdjudicatedCases++
	}
	if passed {
		summary.CasesPassed++
	}
	switch report.Receipt.Classification {
	case ClassPreserved:
		summary.SemanticPreserved++
	case ClassChanged:
		summary.SemanticChanged++
	case ClassIndeterminate:
		summary.Indeterminate++
	}
	if report.Receipt.Resolution == ResolutionUnknown {
		summary.UnknownPaths++
	}
	summary.RepositoryWrites += report.RepositoryWrites
}

func ratio(numerator, denominator int) int {
	if denominator == 0 {
		return 10000
	}
	return numerator * 10000 / denominator
}
