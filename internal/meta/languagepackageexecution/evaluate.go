package languagepackageexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/packageexecution"

func Evaluate(input Input) Report {
	report := Report{Schema: ReportSchema, HeadSHA: input.HeadSHA, Cases: []CaseResult{}, Indicators: []Indicator{}, Proofs: []Proof{}, Views: []AudienceView{}}
	evidence := evidenceByID(input.Cases)
	for _, spec := range input.Contract.Cases {
		result := evaluateCase(spec, evidence[spec.ID])
		report.Cases = append(report.Cases, result)
	}
	report.Summary = summarize(input.Contract, input.Cases, report.Cases)
	bindReplay(&report)
	report.Indicators = indicators(report.Summary)
	report.FactsDigest = digestValue(struct {
		Summary    Summary
		Cases      []CaseResult
		Indicators []Indicator
	}{report.Summary, report.Cases, report.Indicators})
	report.Proofs = proofs(report)
	report.Views = views(report)
	decide(&report)
	seal(&report)
	return report
}

func evidenceByID(values []CaseEvidence) map[string]CaseEvidence {
	result := make(map[string]CaseEvidence, len(values))
	for _, value := range values {
		if _, exists := result[value.ID]; exists {
			result[value.ID] = CaseEvidence{ID: value.ID}
			continue
		}
		result[value.ID] = value
	}
	return result
}

func evaluateCase(spec CaseSpec, evidence CaseEvidence) CaseResult {
	receipt := evidence.Receipt
	result := CaseResult{ID: spec.ID, Decision: receipt.Decision, Reason: receipt.Reason, Resolution: receipt.Resolution, ReceiptDigest: receipt.Digest}
	if evidence.ID == "" {
		result.Decision = "FAIL_CLOSED"
		result.Reason = "CASE_EVIDENCE_MISSING"
		result.Resolution = "EXACT"
		return result
	}
	if receipt.Decision != "PASS" && receipt.Decision != "FAIL_CLOSED" {
		result.Decision = "FAIL_CLOSED"
		result.Reason = "PACKAGE_EXECUTION_DECISION_UNKNOWN"
		result.Resolution = "LOWER_RESOLUTION"
		return result
	}
	if packageexecution.Validate(receipt) != nil {
		result.Decision = "FAIL_CLOSED"
		result.Reason = "CASE_RECEIPT_INVALID"
		result.Resolution = "EXACT"
		return result
	}
	result.Satisfied = receipt.Decision == spec.ExpectedDecision && receipt.Reason == spec.ExpectedReason
	return result
}

func bindReplay(report *Report) {
	if len(report.Cases) < 2 {
		return
	}
	if report.Cases[0].ReceiptDigest == "" || report.Cases[0].ReceiptDigest != report.Cases[1].ReceiptDigest {
		report.Cases[1].Satisfied = false
		report.Cases[1].Decision = "FAIL_CLOSED"
		report.Cases[1].Reason = "DETERMINISTIC_REPLAY_MISMATCH"
		report.Cases[1].Resolution = "EXACT"
	}
	report.Summary.CasesSatisfied = countSatisfied(report.Cases)
	report.Summary.DeterministicReplays = boolInt(report.Cases[1].Satisfied)
}

func decide(report *Report) {
	report.RepositoryWrites = report.Summary.RepositoryWrites
	report.MutationAuthority = report.Summary.MutationAuthorities != 0
	if report.Summary.UnknownDecisions != 0 {
		report.Decision = "FAIL_CLOSED"
		report.Reason = "PACKAGE_EXECUTION_DECISION_UNKNOWN"
		report.Resolution = "LOWER_RESOLUTION"
		return
	}
	if report.Summary.CasesSatisfied == report.Summary.CasesTotal && report.RepositoryWrites == 0 && !report.MutationAuthority {
		report.Decision = "PASS"
		report.Reason = "PACKAGE_EXECUTION_CONTRACT_SATISFIED"
		report.Resolution = "EXACT"
		return
	}
	report.Decision = "FAIL_CLOSED"
	report.Reason = "PACKAGE_EXECUTION_CONTRACT_UNSATISFIED"
	report.Resolution = "EXACT"
}
