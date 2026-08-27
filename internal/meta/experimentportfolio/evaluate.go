package experimentportfolio

import "reflect"

func Evaluate(input Input) Report {
	if reason := inputReason(input); reason != "" {
		return closed(input, reason)
	}
	comparisons := make([]CandidateComparison, 0, len(input.Receipts))
	for _, receipt := range input.Receipts {
		comparisons = append(comparisons, CandidateComparison{
			CandidateID:          receipt.CandidateID,
			SourcePath:           receipt.SourcePath,
			MetaOperation:        receipt.MetaOperation,
			Producer:             receipt.Producer,
			Consumer:             receipt.Consumer,
			ProofChoice:          receipt.ProofChoice,
			Receipt:              receipt,
			CoordinateVector:     append([]Coordinate(nil), receipt.CoordinateVector...),
			CounterexampleCount:  len(receipt.Counterexamples),
			Counterexamples:      append([]Counterexample(nil), receipt.Counterexamples...),
			UnknownLocationCount: len(receipt.UnknownLocations),
			UnknownLocations:     append([]UnknownLocation(nil), receipt.UnknownLocations...),
			ExtensionEvidence:    append([]ExtensionEvidence(nil), receipt.ExtensionEvidence...),
		})
	}
	report := Report{
		Schema:         ReportSchema,
		Decision:       "PORTFOLIO_PRESERVED",
		Resolution:     "EXACT",
		Reason:         "NO_WINNER_NO_AGGREGATE",
		Interpretation: "FIXED_VECTORS_WITH_COUNTEREXAMPLES_AND_UNKNOWN_LOCATIONS",
		SubjectSHA:     input.SubjectSHA,
		ContractID:     input.Contract.ID,
		Candidates:     comparisons,
		Summary:        summarize(comparisons),
		NotClaimed:     append([]string(nil), input.Contract.NotClaimed...),
		FactsDigest:    digestValue(input),
	}
	report.RepositoryWrites = report.Summary.RepositoryWrites
	report.MutationAuthority = report.Summary.MutationAuthority
	report.Proofs = buildProofs(report)
	return sealReport(report)
}

func closed(input Input, reason string) Report {
	notClaimed := ExpectedContract().NotClaimed
	if len(input.Contract.NotClaimed) > 0 {
		notClaimed = input.Contract.NotClaimed
	}
	report := Report{
		Schema:            ReportSchema,
		Decision:          "FAIL_CLOSED",
		Resolution:        "LOWER_RESOLUTION",
		Reason:            reason,
		Interpretation:    "NO_COMPARISON_CLAIM",
		SubjectSHA:        input.SubjectSHA,
		ContractID:        input.Contract.ID,
		Candidates:        []CandidateComparison{},
		Summary:           Summary{Candidates: 0, CoordinatesPerCandidate: ExpectedCoordinates, CounterexampleCounts: map[string]int{}, UnknownLocationIDs: map[string][]string{}, ExtensionEvidenceStatuses: map[string][]string{}, Unknowns: 1},
		Proofs:            []Proof{},
		NotClaimed:        append([]string(nil), notClaimed...),
		RepositoryWrites:  0,
		MutationAuthority: false,
		FactsDigest:       digestValue(input),
	}
	return sealReport(report)
}

func buildProofs(report Report) []Proof {
	preserved := true
	for _, candidate := range report.Candidates {
		if !reflect.DeepEqual(candidate.Receipt.CoordinateVector, candidate.CoordinateVector) ||
			!reflect.DeepEqual(candidate.Receipt.Counterexamples, candidate.Counterexamples) ||
			!reflect.DeepEqual(candidate.Receipt.UnknownLocations, candidate.UnknownLocations) ||
			!reflect.DeepEqual(candidate.Receipt.ExtensionEvidence, candidate.ExtensionEvidence) ||
			candidate.CounterexampleCount != len(candidate.Counterexamples) ||
			candidate.UnknownLocationCount != len(candidate.UnknownLocations) {
			preserved = false
		}
	}
	return []Proof{
		{Choice: "RECEIPT_INTEGRITY", MetaOperation: "seal-independent-receipts", Stage: "FOUNDATION", Step: "verify-digests", Reason: "each producer receipt is content-bound", EvidenceDigest: report.FactsDigest, Passed: len(report.Candidates) == ExpectedCandidates},
		{Choice: "VECTOR_PRESERVATION", MetaOperation: "adjudicate-preserved-vector", Stage: "COMPARISON", Step: "copy-coordinates", Reason: "the evaluator copies every coordinate without collapsing it", EvidenceDigest: report.FactsDigest, Passed: preserved},
		{Choice: "NO_AGGREGATION", MetaOperation: "retain-counterexamples-and-unknowns", Stage: "GOVERNANCE", Step: "block-winner-claim", Reason: "no score, rank, winner, estimated improvement, or weighted average is emitted", EvidenceDigest: report.FactsDigest, Passed: report.Summary.RepositoryWrites == 0 && !report.Summary.MutationAuthority},
	}
}

func summarize(values []CandidateComparison) Summary {
	summary := Summary{
		Candidates:                len(values),
		CounterexampleCounts:      map[string]int{},
		UnknownLocationIDs:        map[string][]string{},
		ExtensionEvidenceStatuses: map[string][]string{},
	}
	for _, candidate := range values {
		if summary.CoordinatesPerCandidate == 0 {
			summary.CoordinatesPerCandidate = len(candidate.CoordinateVector)
		}
		summary.CounterexampleCounts[candidate.CandidateID] = candidate.CounterexampleCount
		unknownIDs := make([]string, 0, len(candidate.UnknownLocations))
		for _, unknown := range candidate.UnknownLocations {
			unknownIDs = append(unknownIDs, unknown.ID)
		}
		summary.UnknownLocationIDs[candidate.CandidateID] = unknownIDs
		statuses := make([]string, 0, len(candidate.ExtensionEvidence))
		for _, evidence := range candidate.ExtensionEvidence {
			statuses = append(statuses, evidence.Status)
		}
		summary.ExtensionEvidenceStatuses[candidate.CandidateID] = statuses
		summary.RepositoryWrites += candidate.Receipt.RepositoryWrites
		summary.MutationAuthority = summary.MutationAuthority || candidate.Receipt.MutationAuthority
	}
	return summary
}
