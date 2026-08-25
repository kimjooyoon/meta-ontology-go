package languagetestexperiment

func buildIndicators(contract Contract, value facts) []Indicator {
	return []Indicator{
		makeIndicator("test.receipts", "OUTCOME", "FOUNDATION", "count-language-test-receipts", value.Receipts, contract.ExpectedReceipts),
		makeIndicator("test.declared", "OUTCOME", "FOUNDATION", "count-declared-language-tests", value.DeclaredTests, contract.ExpectedDeclaredTests),
		makeIndicator("test.executed", "OUTCOME", "COHERENCE", "count-executed-language-tests", value.ExecutedTests, contract.ExpectedExecutedTests),
		makeIndicator("test.passed", "OUTCOME", "COHERENCE", "count-passed-language-tests", value.PassedTests, contract.ExpectedPassedTests),
		makeIndicator("test.source-coherence", "COHERENCE", "COHERENCE", "compare-language-test-sources", value.SourceCoherence, contract.ExpectedSourceCoherence),
		makeIndicator("test.receipt-digest-variants", "REGRESSION", "REGRESSION", "count-language-test-receipt-digests", value.ReceiptDigestVariants, contract.ExpectedReceiptDigestVariants),
		makeIndicator("test.execution-digest-variants", "REGRESSION", "REGRESSION", "count-language-test-execution-digests", value.ExecutionDigestVariants, contract.ExpectedExecutionDigestVariants),
		makeIndicator("compiler.go127-executable", "FOUNDATION", "FOUNDATION", "bind-go127-language-test-runner", value.Go127Runtimes, contract.ExpectedGo127Runtimes),
		makeIndicator("counterexample.assertion-failure", "REGRESSION", "REGRESSION", "count-failed-output-assertions", value.AssertionRejections, contract.ExpectedAssertionRejections),
		makeIndicator("counterexample.missing-tests", "REGRESSION", "REGRESSION", "count-missing-test-rejections", value.MissingTestRejections, contract.ExpectedMissingTestRejections),
		makeIndicator("guardrail.effects", "EFFECT", "FOUNDATION", "sum-language-test-effects", value.RepositoryWrites, contract.ExpectedRepositoryWrites),
		makeIndicator("guardrail.non-claims", "FOUNDATION", "FOUNDATION", "count-language-test-non-claims", value.NonClaims, contract.ExpectedNonClaims),
	}
}

func makeIndicator(id, class, proof, operation string, observed, expected int) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Observed: observed, Expected: expected, Satisfied: observed == expected}
}

func buildViews(indicators []Indicator) []View {
	return []View{
		makeView("USER", "USER_VISIBLE", indicators[:4]),
		makeView("TOOL_AUTHOR", "TOOL_CONTRACT", indicators[:8]),
		makeView("GOVERNOR", "FULL_RECEIPT", indicators),
	}
}

func makeView(audience, resolution string, indicators []Indicator) View {
	view := View{Audience: audience, Resolution: resolution, Total: len(indicators)}
	for _, indicator := range indicators {
		view.IndicatorIDs = append(view.IndicatorIDs, indicator.ID)
		if indicator.Satisfied {
			view.Satisfied++
		}
	}
	view.BasisPoints = basisPoints(view.Satisfied, view.Total)
	return view
}
