package toolchainlsp

import "strings"

func Evaluate(headSHA string, corpus Corpus, concept ConceptBinding) Report {
	report := Report{Schema: ReportSchema, Decision: DecisionFailClosed, Resolution: ResolutionInvariant,
		HeadSHA: headSHA, CorpusDigest: digestValue(corpus), ConceptDigest: concept.ArtifactDigest}
	if !validSHA(headSHA) {
		report.Reason = "TOOLCHAIN_LSP_HEAD_UNKNOWN"; report.Summary.HeadMismatches = 1
		return finish(report)
	}
	if err := ValidateCorpus(corpus); err != nil {
		report.Reason = "TOOLCHAIN_LSP_CORPUS_DRIFT"; report.Summary.CorpusDrift = 1
		return finish(report)
	}
	if reason, err := validateConcept(concept); err != nil {
		report.Reason = reason; report.Summary.ConceptDrift = 1
		return finish(report)
	}
	observations, stats, err := evaluateRuntime()
	if err != nil {
		report.Reason = "TOOLCHAIN_LSP_RUNTIME_UNRESOLVED"; report.Summary.Unresolved = 1
		return finish(report)
	}
	report.Summary = summarize(observations, stats, concept)
	for _, expected := range caseContract {
		observed, ok := observations[expected.ID]
		status := "UNRESOLVED"
		value := "MISSING"
		if ok { value = observed.Observed; if observed.Satisfied { status = "SATISFIED" } }
		report.Cases = append(report.Cases, CaseResult{ID: expected.ID, Group: expected.Group,
			Expected: expected.Expected, Observed: value, Status: status})
	}
	report.Proofs = buildProofs(report.Summary, report.CorpusDigest, report.ConceptDigest)
	for _, proof := range report.Proofs { if !proof.Passed { report.Summary.ProofFailures++ } }
	if report.Summary.CasesSatisfied == len(caseContract) && report.Summary.ProofFailures == 0 {
		report.Decision, report.Reason, report.Resolution = DecisionPass, "TOOLCHAIN_LSP_READY", ResolutionExact
	} else { report.Reason = "TOOLCHAIN_LSP_CASES_UNRESOLVED" }
	report.Indicators = buildIndicators(report.Summary, report.Resolution)
	report.RepositoryWrites = report.Summary.RepositoryWrites
	return finish(report)
}

func validSHA(value string) bool {
	if len(value) != 40 { return false }
	for _, character := range value { if !strings.ContainsRune("0123456789abcdef", character) { return false } }
	return true
}
