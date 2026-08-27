package denominatorevolutionverify

import (
	"bytes"
	"encoding/json"
	"reflect"
)

func Verify(input Input) Verification {
	verification := Verification{Schema: VerificationSchema, HeadSHA: input.HeadSHA, Producer: "denominatorevolution.Evaluate", Consumer: "denominatorevolutionverify.Verify", SourceDigest: input.SourceDigest, NotClaimed: []string{"improvement rate", "aggregate estimate", "projected coverage", "semantic quality", "repository mutation"}}
	var contract Contract
	var report Report
	contractOK := decode(input.ContractRaw, &contract)
	reportOK := decode(input.ReportRaw, &report)
	if contractOK {
		verification.ContractDigest = digestValue(contract)
	} else {
		verification.ContractDigest = digestBytes(input.ContractRaw)
	}
	if reportOK {
		verification.ReportDigest = report.Digest
	}
	checks := []Check{
		verifyCheck("contract-schema", "FOUNDATION", "bind-contract-version", "FOUNDATION", "decode-contract", contractOK && contract.Schema == ContractSchema && contract.Version == 1, "contract v1", contract.Schema),
		verifyCheck("report-schema", "FOUNDATION", "bind-report-version", "FOUNDATION", "decode-report", reportOK && report.Schema == ReportSchema, ReportSchema, report.Schema),
		verifyCheck("report-digest", "REGRESSION", "recompute-report-digest", "REPLAY", "compare-report-digest", reportOK && report.Digest == reportDigest(report), "recomputed digest", report.Digest),
		verifyCheck("contract-digest", "FOUNDATION", "bind-contract-digest", "FOUNDATION", "compare-contract-digest", reportOK && report.ContractDigest == verification.ContractDigest, verification.ContractDigest, report.ContractDigest),
		verifyCheck("source-digest", "FOUNDATION", "bind-source-digest", "FOUNDATION", "compare-source-digest", reportOK && input.SourceDigest != "" && report.SourceDigest == input.SourceDigest, input.SourceDigest, report.SourceDigest),
	}
	if reportOK {
		checks = append(checks, verifyDenominator(report, contract)...)
		checks = append(checks, verifyCases(report, contract)...)
		checks = append(checks, verifySummary(report)...)
	}
	verification.Checks = checks
	verification.RepositoryWrites = 0
	verification.MutationAuthority = false
	verification.Decision, verification.Resolution, verification.Reason = "PASS", "EXACT", "INDEPENDENT_DENOMINATOR_DECISION_VERIFIED"
	if !contractOK || !reportOK || hasFailure(checks) || report.HeadSHA != input.HeadSHA || report.Decision != "PASS" || report.Resolution != "EXACT" {
		verification.Decision, verification.Resolution, verification.Reason = "FAIL_CLOSED", "INVARIANT_ONLY", "INDEPENDENT_DENOMINATOR_DECISION_REJECTED"
	}
	verification.Digest = verificationDigest(verification)
	return verification
}

func decode(raw []byte, value any) bool {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value) == nil
}

func verifyDenominator(report Report, contract Contract) []Check {
	denominatorOK := report.Denominator.Version == DenominatorVersion && len(report.Denominator.Obligations) == DenominatorSize && report.Denominator.Digest == denominatorDigest(report.Denominator)
	contractOK := contract.Denominator.Version == DenominatorVersion && len(contract.Denominator.Obligations) == DenominatorSize && reflect.DeepEqual(report.Denominator.Obligations, contract.Denominator.Obligations)
	return []Check{
		verifyCheck("fixed-denominator", "FOUNDATION", "bind-fixed-denominator", "FOUNDATION", "compare-member-set", denominatorOK && contractOK, "5 members at "+DenominatorVersion, report.Denominator.Version+" / "+report.Denominator.Digest),
		verifyCheck("no-composite-estimate", "REGRESSION", "reject-aggregate-estimate", "GUARD", "reject-forbidden-claim", report.Summary.ForbiddenEstimateNumerator == 0 && report.Summary.ForbiddenEstimateDenominator == 1, "0/1", intText(report.Summary.ForbiddenEstimateNumerator)+"/"+intText(report.Summary.ForbiddenEstimateDenominator)),
	}
}

func verifyCases(report Report, contract Contract) []Check {
	if len(report.Cases) != CaseCount || len(contract.Cases) != CaseCount {
		return []Check{verifyCheck("case-cardinality", "FOUNDATION", "bind-case-corpus", "FOUNDATION", "compare-case-count", false, "3 cases", intText(len(report.Cases)))}
	}
	checks := []Check{}
	for index, value := range report.Cases {
		spec := contract.Cases[index]
		basic := value.ID == spec.ID && value.Kind == spec.Kind && value.Status == "SATISFIED" && value.ExpectedDecision == spec.ExpectedDecision && value.ExpectedResolution == spec.ExpectedResolution && value.ExpectedReason == spec.ExpectedReason && value.ObservedDecision == spec.ExpectedDecision && value.ObservedResolution == spec.ExpectedResolution && value.ObservedReason == spec.ExpectedReason && value.FromClaim == spec.FromClaim && value.ToClaim == spec.ToClaim && len(value.Checks) == 8
		checks = append(checks, verifyCheck(value.ID, spec.ProofChoice, spec.MetaOperation, spec.Stage, spec.Step, basic, spec.ExpectedDecision+" / "+spec.ExpectedReason, value.ObservedDecision+" / "+value.ObservedReason))
		if value.ID == "legal-advance" {
			checks = append(checks, verifyLegalReceipt(value, spec)...)
		}
		if value.ID == "unauthorized-change" {
			checks = append(checks, verifyUnauthorized(value, spec)...)
		}
		if value.ID == "unknown-predecessor" {
			checks = append(checks, verifyUnknown(value, spec)...)
		}
	}
	return checks
}

func verifyLegalReceipt(value Case, spec CaseSpec) []Check {
	receipt := value.Receipt
	ok := receipt != nil && receipt.Schema == ReceiptSchema && receipt.Decision == "ADVANCE" && receipt.Reason == "DENOMINATOR_ADVANCE_AUTHORIZED" && receipt.RepositoryWrites == 0 && !receipt.MutationAuthority && receipt.Predecessor.Version == value.Predecessor.Version && receipt.Predecessor.Digest == value.Predecessor.Digest && receipt.Successor.Version == value.Successor.Version && receipt.Successor.Digest == value.Successor.Digest && receipt.Digest == receiptDigest(*receipt) && len(receipt.Additions) == 1 && receipt.Additions[0].ObligationID == "governance/replay-receipt" && receipt.Additions[0].Reason == "NEW_MEASURABLE_OBLIGATION" && len(receipt.Deletions) == 1 && receipt.Deletions[0].ObligationID == "governance/legacy-summary" && receipt.Deletions[0].Reason == "DEPRECATED_DUPLICATE"
	return []Check{verifyCheck("legal-migration-receipt", spec.ProofChoice, "accept-migration-receipt", "MIGRATE", "replay-receipt", ok, "bound v1 -> v2 receipt", receiptText(receipt))}
}

func verifyUnauthorized(value Case, spec CaseSpec) []Check {
	ok := value.Receipt == nil && value.ObservedDecision == "BLOCK" && value.ObservedReason == "MIGRATION_RECEIPT_MISSING"
	return []Check{verifyCheck("unauthorized-no-receipt", spec.ProofChoice, "reject-unreceipted-denominator-change", "DECIDE", "reject-missing-receipt", ok, "no receipt and BLOCK", value.ObservedDecision+" / "+receiptText(value.Receipt))}
}

func verifyUnknown(value Case, spec CaseSpec) []Check {
	ok := value.Receipt == nil && value.Predecessor.Version != DenominatorVersion && value.ObservedDecision == "FAIL_CLOSED" && value.ObservedReason == "PREDECESSOR_DIGEST_UNKNOWN"
	return []Check{verifyCheck("unknown-predecessor", spec.ProofChoice, "fail-closed-unknown-predecessor", "RESOLVE", "lookup-predecessor", ok, "unknown predecessor stays UNKNOWN", value.ObservedDecision+" / "+value.ObservedReason)}
}

func verifySummary(report Report) []Check {
	summary := report.Summary
	ok := summary.CasesSatisfied == 3 && summary.CasesTotal == 3 && summary.FixedDenominatorNumerator == 5 && summary.FixedDenominatorDenominator == 5 && summary.LegalAdvanceNumerator == 1 && summary.LegalAdvanceDenominator == 1 && summary.UnauthorizedRejectionNumerator == 1 && summary.UnauthorizedRejectionDenominator == 1 && summary.UnknownPredecessorNumerator == 1 && summary.UnknownPredecessorDenominator == 1 && summary.AdditionReasonNumerator == 1 && summary.AdditionReasonDenominator == 1 && summary.DeletionReasonNumerator == 1 && summary.DeletionReasonDenominator == 1 && summary.ForbiddenEstimateNumerator == 0 && summary.ForbiddenEstimateDenominator == 1 && summary.RepositoryWrites == 0 && summary.MutationAuthorities == 0
	return []Check{verifyCheck("exact-summary", "FOUNDATION", "bind-exact-counters", "SUMMARY", "compare-numerators-and-denominators", ok, "fixed 5/5 and three separate 1/1 predicates", "summary counters")}
}
