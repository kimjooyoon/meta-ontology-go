package denominatorevolution

import "fmt"

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.Scope != ReportScope || report.HeadSHA == "" || report.Producer == "" || report.Consumer == "" {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_IDENTITY_MISMATCH")
	}
	if report.Decision != "PASS" || report.Resolution != "EXACT" || report.Reason != "DENOMINATOR_EVOLUTION_CONTRACT_SATISFIED" {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_DECISION_MISMATCH")
	}
	if report.Denominator.Version != DenominatorVersion || len(report.Denominator.Obligations) != DenominatorSize || report.Denominator.Digest != denominatorDigest(report.Denominator) || report.SourceProjection.ObligationCount != DenominatorSize || report.SourceProjection.CaseCount != CaseCount {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_DENOMINATOR_MISMATCH")
	}
	if len(report.Cases) != CaseCount || len(report.Indicators) != CheckCount || len(report.ClaimLedger) != CaseCount || report.Summary.CasesSatisfied != CaseCount || report.Summary.CasesTotal != CaseCount || report.Summary.SourceCasesNumerator != CaseCount || report.Summary.SourceCasesDenominator != CaseCount || report.Summary.PersistentClaimsNumerator != CaseCount || report.Summary.PersistentClaimsDenominator != CaseCount || report.Summary.GuardrailObservationsNumerator != GuardrailCount || report.Summary.GuardrailObservationsDenominator != GuardrailCount {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_CARDINALITY_MISMATCH")
	}
	for _, value := range report.Cases {
		if value.Status != "SATISFIED" || len(value.Checks) != CheckCount {
			return fmt.Errorf("DENOMINATOR_EVOLUTION_CASE_MISMATCH")
		}
	}
	if hasUnsatisfied(report.Indicators) || !guardrailsConform(report.Summary.Guardrails) || report.RepositoryWrites != report.RepositorySnapshot.ChangedPaths || report.RepositoryWrites != 0 || report.MutationAuthority || report.Digest != reportDigest(report) {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_EFFECT_OR_DIGEST_MISMATCH")
	}
	return nil
}
