package main

import producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"

const (
	contractReproduced            = "FIXED_FIVE_CASE_CONTRACT_REPRODUCED"
	contractIncomplete            = "FIXED_FIVE_CASE_CONTRACT_INCOMPLETE"
	subjectEquivalenceNotAsserted = "NOT_ASSERTED"
)

func runSuite(subjectSHA, observedCheckoutSHA, effectsBefore, effectsAfter, outputPath string) Suite {
	definitions := producer.Denominator()
	meta, metaErr := producer.ReadMetaContract()
	caseContractOK := caseContractValid(definitions, meta, metaErr)
	results := make([]CaseResult, 0, fixedCaseContractTotal)
	summary := Summary{CasesTotal: fixedCaseContractTotal, ModeledSemanticComponents: producer.ModeledComponentCount, TotalSemanticComponents: producer.TotalComponentCount}
	passed := 0
	for _, definition := range definitions {
		input := producer.Input{CaseID: definition.ID, SubjectSHA: subjectSHA, ObservedCheckoutSHA: observedCheckoutSHA, BeforePath: definition.BeforePath, AfterPath: definition.AfterPath, EffectsBeforePath: effectsBefore, EffectsAfterPath: effectsAfter, OutputPath: outputPath}
		report := evaluate(input, "")
		semanticCoverageAccepted := report.Receipt.DeclaredProjectionComponentKindCoverageBPS == 10000 || definition.ExpectedResolution == producer.ResolutionLower
		claimIdentityExact := report.ClaimIdentity.Decision == "EXACT" && report.ClaimIdentity.Resolution == producer.ResolutionExact && report.ClaimIdentity.CaseID == definition.ID && report.ClaimIdentity.FixedTotal == len(report.ClaimIdentity.ExpectedClaimIDs) && report.ClaimIdentity.CoverageBPS == 10000
		casePassed := caseContractOK && claimIdentityExact && report.IndependentVerdict.Passed && report.IndependentVerdict.Decision == definition.ExpectedDecision && report.IndependentVerdict.Resolution == definition.ExpectedResolution && report.IndependentVerdict.Classification == definition.ExpectedClass && report.IndependentVerdict.Reason == definition.ExpectedReason && report.Receipt.Stage == definition.ExpectedStage && report.Receipt.Step == definition.ExpectedStep && report.Receipt.DenominatorVersion == producer.DenominatorVersion && report.Receipt.DenominatorCases == fixedCaseContractTotal && semanticCoverageAccepted
		if casePassed {
			passed++
		}
		results = append(results, CaseResult{Definition: definition, Passed: casePassed, Report: report})
		mergeSummary(&summary, report, casePassed)
	}
	summary.ClaimStatusCoverageBPS = ratio(summary.ClaimsWithExplainedStatus, summary.TotalClaims)
	contract := contractIncomplete
	decision, resolution := producer.DecisionFailClosed, producer.ResolutionInvariant
	reason := ""
	if !caseContractOK {
		if metaErr != nil {
			reason = caseContractMetaErrorReason
		} else {
			reason = caseContractMismatchReason
		}
	}
	if caseContractOK && meta.Version == producer.DenominatorVersion && meta.DenominatorCases == fixedCaseContractTotal && passed == fixedCaseContractTotal {
		decision, resolution, contract = producer.DecisionFixedPoint, producer.ResolutionExact, contractReproduced
		reason = ""
	}
	sources := []string{producer.MetaSourcePath}
	for _, definition := range definitions {
		sources = append(sources, definition.BeforePath, definition.AfterPath)
	}
	denominatorDigest := digestValue(fixedCaseRecipes)
	metaDigest := meta.Digest
	if metaErr != nil {
		metaDigest = ""
	}
	caseContractReason := caseContractExactReason
	if !caseContractOK {
		caseContractReason = reason
	}
	suite := Suite{Schema: producer.SuiteSchema, SubjectSHA: subjectSHA, ObservedCheckoutSHA: observedCheckoutSHA, DenominatorID: fixedCaseContractDenominatorID, DenominatorDigest: denominatorDigest, CaseContractDenominatorID: fixedCaseContractDenominatorID, CaseContractExpectedIDs: fixedCaseIDs(), CaseContractObservedIDs: observedCaseIDs(definitions), CaseContractObservedRecipeIDs: observedRecipeIDs(meta), CaseContractFixedTotal: fixedCaseContractTotal, CaseContractStage: caseContractStage, CaseContractStep: caseContractStep, CaseContractReason: caseContractReason, Stage: caseContractStage, Step: caseContractStep, Decision: decision, Resolution: resolution, Reason: reason, ContractReproduction: contract, SubjectSemanticEquivalence: subjectEquivalenceNotAsserted, SourcePaths: sources, OutputPath: outputPath, Cases: results, Summary: summary, CaseContractCoverageBPS: ratio(passed, fixedCaseContractTotal), DeclaredProjectionComponentKindCoverageBPS: ratio(producer.ModeledComponentCount, producer.TotalComponentCount), SemanticEquivalenceClaim: subjectEquivalenceNotAsserted, MetaSourcePath: producer.MetaSourcePath, MetaContractDigest: metaDigest, DenominatorVersion: producer.DenominatorVersion, ModeledSemanticComponents: producer.ModeledComponentCount, TotalSemanticComponents: producer.TotalComponentCount}
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
	summary.ChangedPathOrContentCount += report.Receipt.Effects.ChangedPathOrContentCount
	summary.ClaimsWithExplainedStatus += local.ClaimsWithExplainedStatus
	summary.TotalClaims += local.TotalClaims
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
