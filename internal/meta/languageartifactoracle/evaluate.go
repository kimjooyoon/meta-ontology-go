package languageartifactoracle

import "reflect"

func Evaluate(input Input) Report {
	if !reflect.DeepEqual(input.Contract, CanonicalContract()) {
		return failedReport(input.HeadSHA, "ARTIFACT_ORACLE_CONTRACT_DRIFT")
	}
	cases := evaluateCases(input)
	summary := summarize(input, cases)
	decision, resolution, reason := "PASS", "EXACT", "ARTIFACT_ORACLE_CONTRACT_SATISFIED"
	if summary.CasesSatisfied != CaseTotal || summary.ProducerDependencies != 0 ||
		summary.LegacyValidatorCounterexamples != 1 || summary.SemanticCorrectnessClaims != 0 {
		decision, resolution, reason = "FAIL_CLOSED", "INVARIANT_ONLY", "ARTIFACT_ORACLE_CONTRACT_VIOLATED"
	}
	report := Report{Schema: ReportSchema, Scope: ReportScope, HeadSHA: input.HeadSHA,
		Decision: decision, Resolution: resolution, Reason: reason,
		ContractDigest: digestValue(input.Contract), IndependenceDigest: digestValue(input.Independence),
		LegacyDigest: digestBytes(input.LegacyAcceptance), Cases: cases, Summary: summary,
		Indicators: indicators(summary), NotClaimed: []string{
			"full compiler semantic correctness", "value-level computation", "external side effects",
			"general parsing outside the fixed oracle grammar",
		}, RepositoryWrites: 0, MutationAuthority: false}
	report.Digest = reportDigest(report)
	return report
}

func failedReport(head, reason string) Report {
	report := Report{Schema: ReportSchema, Scope: ReportScope, HeadSHA: head,
		Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Reason: reason,
		Summary: Summary{CasesTotal: CaseTotal, UnknownChecks: CaseTotal * CheckTotal},
		Cases:   []CaseResult{}, Indicators: []Indicator{},
		NotClaimed: []string{"artifact binding without a canonical oracle contract"}}
	report.Digest = reportDigest(report)
	return report
}
