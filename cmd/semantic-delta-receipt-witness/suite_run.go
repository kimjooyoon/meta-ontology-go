package main

import producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"

const (
	contractReproduced            = "FIXED_THREE_CASE_CONTRACT_REPRODUCED"
	subjectEquivalenceNotAsserted = "NOT_ASSERTED"
)

func runSuite(subjectSHA, outputPath string) Suite {
	definitions := producer.Denominator()
	results := make([]CaseResult, 0, len(definitions))
	summary := Summary{CasesTotal: len(definitions)}
	passed := 0
	for _, definition := range definitions {
		input := producer.Input{CaseID: definition.ID, SubjectSHA: subjectSHA, BeforePath: definition.BeforePath, AfterPath: definition.AfterPath}
		report := evaluate(input, "")
		casePassed := report.IndependentVerdict.Passed && report.IndependentVerdict.Decision == definition.ExpectedDecision && report.IndependentVerdict.Resolution == definition.ExpectedResolution && report.IndependentVerdict.Classification == definition.ExpectedClass && report.IndependentVerdict.Reason == definition.ExpectedReason
		if casePassed {
			passed++
		}
		results = append(results, CaseResult{Definition: definition, Passed: casePassed, Report: report})
		mergeSummary(&summary, report, casePassed)
	}
	decision, resolution := producer.DecisionFailClosed, producer.ResolutionInvariant
	contract := "FIXED_THREE_CASE_CONTRACT_INCOMPLETE"
	if passed == len(definitions) {
		decision, resolution = producer.DecisionFixedPoint, producer.ResolutionExact
		contract = contractReproduced
	}
	summary.CasesPassed = passed
	sources := make([]string, 0, len(definitions)*2)
	for _, definition := range definitions {
		sources = append(sources, definition.BeforePath, definition.AfterPath)
	}
	suite := Suite{Schema: producer.SuiteSchema, SubjectSHA: subjectSHA, DenominatorID: producer.DenominatorID, DenominatorDigest: digestValue(definitions), Decision: decision, Resolution: resolution, ContractReproduction: contract, SubjectSemanticEquivalence: subjectEquivalenceNotAsserted, SourcePaths: sources, OutputPath: outputPath, Cases: results, Summary: summary, CoverageBPS: ratio(passed, len(definitions))}
	sealSuite(&suite)
	return suite
}

func mergeSummary(summary *Summary, report Report, passed bool) {
	local := summaryFor(report.Receipt, report.IndependentVerdict)
	summary.TextualChanges += local.TextualChanges
	summary.StructuralObservations += local.StructuralObservations
	summary.ClaimTransitionCases += local.ClaimTransitionCases
	summary.AdjudicatedCases += local.AdjudicatedCases
	summary.SemanticPreserved += local.SemanticPreserved
	summary.SemanticChanged += local.SemanticChanged
	summary.Indeterminate += local.Indeterminate
	summary.UnknownPaths += local.UnknownPaths
	summary.RepositoryWrites += local.RepositoryWrites + report.RepositoryWrites
	if passed {
		summary.CasesPassed++
	}
}

func ratio(numerator, denominator int) int {
	if denominator == 0 {
		return 10000
	}
	return numerator * 10000 / denominator
}
