package denominatorevolution

import "fmt"

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.Scope != ReportScope || report.HeadSHA == "" || report.Producer == "" || report.Consumer == "" {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_IDENTITY_MISMATCH")
	}
	if report.Decision != "PASS" || report.Resolution != "EXACT" || report.Reason != "DENOMINATOR_EVOLUTION_CONTRACT_SATISFIED" {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_DECISION_MISMATCH")
	}
	if report.Denominator.Version != DenominatorVersion || len(report.Denominator.Obligations) != DenominatorSize || report.Denominator.Digest != denominatorDigest(report.Denominator) || report.SourceProjection.ObligationCount != DenominatorSize || report.SourceProjection.CaseCount != CaseCount || !report.SourceProjection.ForbiddenPropositionPresent {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_DENOMINATOR_MISMATCH")
	}
	if len(report.Cases) != CaseCount || len(report.Indicators) != CheckCount || len(report.ClaimLedger) != CaseCount || len(report.DenominatorRecords) != 2 || report.Summary.CasesSatisfied != CaseCount || report.Summary.CasesTotal != CaseCount || report.Summary.SourceCasesNumerator != CaseCount || report.Summary.SourceCasesDenominator != CaseCount || report.Summary.PersistentClaimsNumerator != CaseCount || report.Summary.PersistentClaimsDenominator != CaseCount || report.Summary.GuardrailObservationsNumerator != GuardrailCount || report.Summary.GuardrailObservationsDenominator != GuardrailCount || report.Summary.VersionRecordsNumerator != 2 || report.Summary.VersionRecordsDenominator != 2 || report.Summary.V1NonretroactiveNumerator != 1 || report.Summary.V1NonretroactiveDenominator != 1 {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_CARDINALITY_MISMATCH")
	}
	if len(report.AggregateMetrics) != 0 || !validDenominatorRecords(report.DenominatorRecords, report.Denominator) || !validClaimLedger(report.ClaimLedger) {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_IMMUTABLE_RECORD_MISMATCH")
	}
	for _, value := range report.Cases {
		if value.Status != "SATISFIED" || len(value.Checks) != CheckCount {
			return fmt.Errorf("DENOMINATOR_EVOLUTION_CASE_MISMATCH")
		}
		if value.Receipt != nil && value.Receipt.Digest != receiptDigest(*value.Receipt) {
			return fmt.Errorf("DENOMINATOR_EVOLUTION_RECEIPT_DIGEST_MISMATCH")
		}
	}
	if hasUnsatisfied(report.Indicators) || !guardrailsConform(report.Summary.Guardrails) || report.RepositoryWrites != report.RepositorySnapshot.ChangedPaths || report.RepositoryWrites != 0 || report.MutationAuthority || report.Digest != reportDigest(report) {
		return fmt.Errorf("DENOMINATOR_EVOLUTION_EFFECT_OR_DIGEST_MISMATCH")
	}
	return nil
}

func validDenominatorRecords(values []DenominatorRecord, base Denominator) bool {
	if len(values) != 2 || values[0].ID != "denominator-v1" || values[0].Version != DenominatorVersion || values[0].Predecessor != nil || !values[0].Immutable || values[0].FixedMemberNumerator != DenominatorSize || values[0].FixedMemberDenominator != DenominatorSize || !reflectDenominator(values[0].Denominator, base) || values[0].Digest != denominatorRecordDigest(values[0]) {
		return false
	}
	if values[1].ID != "denominator-v2" || values[1].Version != SuccessorVersion || values[1].Predecessor == nil || *values[1].Predecessor != (DenominatorRef{Version: base.Version, Digest: base.Digest}) || !values[1].Immutable || values[1].FixedMemberNumerator != DenominatorSize || values[1].FixedMemberDenominator != DenominatorSize || values[1].Denominator.Digest != denominatorDigest(values[1].Denominator) || values[1].Digest != denominatorRecordDigest(values[1]) {
		return false
	}
	return true
}

func validClaimLedger(values []ClaimLedgerEntry) bool {
	previous := ""
	for index, value := range values {
		if value.Sequence != index+1 || value.PriorState != "OPEN" || value.PreviousDigest != previous || value.Digest != claimLedgerDigest(value) {
			return false
		}
		previous = value.Digest
	}
	return true
}
