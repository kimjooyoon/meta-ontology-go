package languageartifactoracle

import "fmt"

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.Scope != ReportScope || report.HeadSHA == "" {
		return fmt.Errorf("ARTIFACT_ORACLE_IDENTITY_MISMATCH")
	}
	if report.Decision != "PASS" || report.Resolution != "EXACT" ||
		report.Reason != "ARTIFACT_ORACLE_CONTRACT_SATISFIED" {
		return fmt.Errorf("ARTIFACT_ORACLE_DECISION_MISMATCH")
	}
	if len(report.Cases) != CaseTotal || len(report.Indicators) != len(MetricIDs()) {
		return fmt.Errorf("ARTIFACT_ORACLE_DENOMINATOR_MISMATCH")
	}
	summary := report.Summary
	if summary.CasesSatisfied != CaseTotal || summary.CasesTotal != CaseTotal ||
		summary.ExactSourceBindings != 1 || summary.ResealedForgeriesRejected != 1 ||
		summary.UnknownDecisionsRejected != 1 || summary.LowerResolutions != 1 ||
		summary.LegacyValidatorCounterexamples != 1 || summary.ProducerDependencies != 0 ||
		summary.UnknownChecks != CheckTotal || summary.SemanticCorrectnessClaims != 0 {
		return fmt.Errorf("ARTIFACT_ORACLE_SUMMARY_MISMATCH")
	}
	for _, result := range report.Cases {
		if result.Status != "SATISFIED" || len(result.Checks) != CheckTotal {
			return fmt.Errorf("ARTIFACT_ORACLE_CASE_MISMATCH")
		}
	}
	for _, indicator := range report.Indicators {
		if !indicator.Satisfied { return fmt.Errorf("ARTIFACT_ORACLE_INDICATOR_MISMATCH") }
	}
	if report.RepositoryWrites != 0 || report.MutationAuthority || report.Digest != reportDigest(report) {
		return fmt.Errorf("ARTIFACT_ORACLE_EFFECT_OR_DIGEST_MISMATCH")
	}
	return nil
}
