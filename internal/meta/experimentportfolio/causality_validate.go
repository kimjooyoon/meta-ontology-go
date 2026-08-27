package experimentportfolio

func causalityInputReason(input CausalityInput) string {
	if reason := contractReason(input.Contract); reason != "" {
		return reason
	}
	if !validSubjectSHA(input.SubjectSHA) {
		return "CAUSALITY_SUBJECT_SHA_INVALID"
	}
	if reason := causalityManifestReason(input.Contract, input.Manifest); reason != "" {
		return reason
	}
	if len(input.Samples) != len(input.Contract.Candidates) {
		return "CAUSALITY_SAMPLE_COUNT_INVALID"
	}
	seen := make(map[string]bool, len(input.Samples))
	for _, sample := range input.Samples {
		if seen[sample.CandidateID] {
			return "CAUSALITY_CANDIDATE_DUPLICATE"
		}
		seen[sample.CandidateID] = true
		candidateCase, ok := causalityCaseContract(input.Manifest, sample.CandidateID)
		if !ok {
			return "CAUSALITY_CANDIDATE_UNKNOWN"
		}
		if reason := causalityCaseReason(input.Contract, input.SubjectSHA, sample.CandidateID, candidateCase, sample.Baseline, "BASELINE", sample.CandidateID+"-baseline", candidateCase.OperationValueBefore); reason != "" {
			return reason
		}
		if reason := causalityCaseReason(input.Contract, input.SubjectSHA, sample.CandidateID, candidateCase, sample.Semantic, "SEMANTIC", sample.CandidateID+"-semantic", candidateCase.OperationValueAfter); reason != "" {
			return reason
		}
		if reason := causalityCaseReason(input.Contract, input.SubjectSHA, sample.CandidateID, candidateCase, sample.NonSemantic, "NON_SEMANTIC", sample.CandidateID+"-nonsemantic", candidateCase.OperationValueBefore); reason != "" {
			return reason
		}
	}
	for _, candidate := range input.Contract.Candidates {
		if !seen[candidate.ID] {
			return "CAUSALITY_CANDIDATE_MISSING"
		}
	}
	return ""
}

func causalityCaseReason(contract Contract, subjectSHA, candidateID string, candidateCase CausalityCaseContract, input CausalityCaseInput, expectedKind, expectedID, expectedValue string) string {
	if input.Kind != expectedKind || input.CaseID != expectedID {
		return "CAUSALITY_CASE_IDENTITY_INVALID"
	}
	if input.Observation.SourcePath != candidateCase.SourcePath || input.Observation.SemanticValue != expectedValue {
		return "CAUSALITY_SOURCE_SEMANTIC_VALUE_MISMATCH"
	}
	if input.Observation.SourceDigest != input.Receipt.SourceDigest {
		return "CAUSALITY_SOURCE_RECEIPT_DIGEST_MISMATCH"
	}
	if reason := receiptReason(contract, subjectSHA, input.Receipt); reason != "" {
		return reason
	}
	if input.Receipt.CandidateID != candidateID {
		return "CAUSALITY_RECEIPT_CANDIDATE_MISMATCH"
	}
	return ""
}

func closedCausality(input CausalityInput, reason string) CausalityReport {
	notClaimed := ExpectedContract().NotClaimed
	if len(input.Contract.NotClaimed) > 0 {
		notClaimed = input.Contract.NotClaimed
	}
	report := CausalityReport{
		Schema:         CausalityReportSchema,
		Decision:       "FAIL_CLOSED",
		Resolution:     "LOWER_RESOLUTION",
		Reason:         reason,
		Interpretation: "NO_CAUSALITY_COMPARISON_CLAIM",
		SubjectSHA:     input.SubjectSHA,
		ContractID:     input.Contract.ID,
		Manifest:       input.Manifest,
		Samples:        []CausalitySampleResult{},
		Summary: CausalitySummary{
			CausalCases: CausalCaseCount{Observed: 0, Total: ExpectedCandidates},
			Unknowns:    1,
		},
		UnknownFindings: []CausalityUnknown{{Stage: "CAUSALITY_VALIDATION", Step: "validate-input", Reason: reason}},
		NotClaimed:      append([]string(nil), notClaimed...),
		FactsDigest:     digestValue(input),
	}
	return sealCausalityReport(report)
}
