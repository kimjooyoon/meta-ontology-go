package main

import producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"

const (
	contractReproduced            = "FIXED_FIVE_CASE_CONTRACT_REPRODUCED"
	contractIncomplete            = "FIXED_FIVE_CASE_CONTRACT_INCOMPLETE"
	subjectEquivalenceNotAsserted = "NOT_ASSERTED"
)

func runSuite(subjectSHA, outputPath string) Suite {
	definitions := producer.Denominator()
	meta, metaErr := producer.ReadMetaContract()
	results := make([]CaseResult, 0, len(definitions))
	summary := Summary{CasesTotal: len(definitions), ModeledSemanticComponents: producer.ModeledComponentCount, TotalSemanticComponents: producer.TotalComponentCount}
	passed := 0
	for _, definition := range definitions {
		input := producer.Input{CaseID: definition.ID, SubjectSHA: subjectSHA, BeforePath: definition.BeforePath, AfterPath: definition.AfterPath}
		report := evaluate(input, "")
		semanticCoverageAccepted := report.Receipt.SemanticCoverageBPS == 10000 || definition.ExpectedResolution == producer.ResolutionLower
		casePassed := metaErr == nil && report.IndependentVerdict.Passed && report.IndependentVerdict.Decision == definition.ExpectedDecision && report.IndependentVerdict.Resolution == definition.ExpectedResolution && report.IndependentVerdict.Classification == definition.ExpectedClass && report.IndependentVerdict.Reason == definition.ExpectedReason && report.IndependentVerdict.Stage == definition.ExpectedStage && report.IndependentVerdict.Step == definition.ExpectedStep && report.Receipt.DenominatorVersion == producer.DenominatorVersion && report.Receipt.DenominatorCases == len(definitions) && semanticCoverageAccepted
		if casePassed {
			passed++
		}
		results = append(results, CaseResult{Definition: definition, Passed: casePassed, Report: report})
		mergeSummary(&summary, report, casePassed)
	}
	contract := contractIncomplete
	decision, resolution := producer.DecisionFailClosed, producer.ResolutionInvariant
	if metaErr == nil && meta.Version == producer.DenominatorVersion && meta.DenominatorCases == len(definitions) && passed == len(definitions) && len(definitions) == 5 {
		decision, resolution, contract = producer.DecisionFixedPoint, producer.ResolutionExact, contractReproduced
	}
	sources := []string{producer.MetaSourcePath}
	for _, definition := range definitions {
		sources = append(sources, definition.BeforePath, definition.AfterPath)
	}
	denominatorDigest := digestValue(definitions)
	metaDigest := meta.Digest
	if metaErr != nil {
		metaDigest = ""
	}
	suite := Suite{Schema: producer.SuiteSchema, SubjectSHA: subjectSHA, DenominatorID: producer.DenominatorID, DenominatorDigest: denominatorDigest, Decision: decision, Resolution: resolution, ContractReproduction: contract, SubjectSemanticEquivalence: subjectEquivalenceNotAsserted, SourcePaths: sources, OutputPath: outputPath, Cases: results, Summary: summary, CoverageBPS: ratio(passed, len(definitions)), MetaSourcePath: producer.MetaSourcePath, MetaContractDigest: metaDigest, DenominatorVersion: producer.DenominatorVersion, ModeledSemanticComponents: producer.ModeledComponentCount, TotalSemanticComponents: producer.TotalComponentCount}
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
	summary.DistinctPropositions += local.DistinctPropositions
	summary.AddedClaims += local.AddedClaims
	summary.RemovedClaims += local.RemovedClaims
	summary.ChangedClaims += local.ChangedClaims
	summary.OpenClaims += local.OpenClaims
	summary.DischargedClaims += local.DischargedClaims
	summary.RefutedClaims += local.RefutedClaims
	summary.TransitionChains += local.TransitionChains
	summary.AmbiguousCases += local.AmbiguousCases
	if passed {
		summary.CasesPassed++
	}
}

func ratio(numerator, denominator int) int {
	if denominator <= 0 {
		return 0
	}
	return numerator * 10000 / denominator
}
