package main

func reduce(expectedSHA string, density densityReport, extraction extractionReport, split splitReport,
	observed, untracked []string) evidence {
	report := evidence{Schema: evidenceSchema, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
		Reason: "WRITE_SET_EVIDENCE_UNKNOWN", Audience: "GOVERNOR", SourceSHA: expectedSHA,
		MetaOperation: metaOperation, Coordinates: coordinates{SourceReceiptsTotal: 3}}
	if issue, receipts := sourceIssue(expectedSHA, density, extraction, split); issue != "" {
		report.Reason, report.Coordinates.SourceReceipts = issue, receipts
		report.Coordinates.Unknowns = 3 - receipts
		return seal(report)
	}
	report.Coordinates.SourceReceipts = 3
	densitySet, densityExact := densityPaths(density)
	if !densityExact {
		report.Reason, report.Coordinates.SourceReceipts, report.Coordinates.Unknowns = "DENSITY_PATH_UNKNOWN", 2, 1
		return seal(report)
	}
	extractionSet, createdSet, extractionExact := extractionPaths(extraction)
	if !extractionExact {
		report.Reason, report.Coordinates.SourceReceipts, report.Coordinates.Unknowns = "EXTRACTION_PATH_UNKNOWN", 2, 1
		return seal(report)
	}
	splitSet, splitCreated, splitExact := splitPaths(split)
	if !splitExact {
		report.Reason, report.Coordinates.SourceReceipts, report.Coordinates.Unknowns = "SPLIT_PATH_UNKNOWN", 2, 1
		return seal(report)
	}
	observedSet, observedExact := pathSet(observed)
	untrackedSet, untrackedExact := pathSet(untracked)
	if !observedExact || !untrackedExact {
		report.Reason, report.Coordinates.Unknowns = "OBSERVED_WRITE_SET_UNKNOWN", 1
		return seal(report)
	}
	expectedSet := unionPaths(unionPaths(densitySet, extractionSet), splitSet)
	createdSet = unionPaths(createdSet, splitCreated)
	report.Expected, report.Observed = sortedPaths(expectedSet), sortedPaths(observedSet)
	report.ExpectedCreated, report.ObservedCreated = sortedPaths(createdSet), sortedPaths(untrackedSet)
	report.Coordinates.DensityPaths, report.Coordinates.ExtractionPaths = len(densitySet), len(extractionSet)
	report.Coordinates.SplitPaths = len(splitSet)
	report.Coordinates.OverlapPaths = len(densitySet) + len(extractionSet) + len(splitSet) - len(expectedSet)
	report.Coordinates.ExpectedPaths, report.Coordinates.ObservedPaths = len(report.Expected), len(report.Observed)
	report.Coordinates.CreatedPaths, report.Coordinates.UntrackedPaths = len(createdSet), len(untrackedSet)
	report.Coordinates.UnclassifiedPaths = mismatchCount(createdSet, untrackedSet)
	report.Resolution = "EXACT"
	if len(extraction.Unhandled) != 0 {
		report.Reason = "EXTRACTION_RESIDUAL_PRESENT"
		return seal(report)
	}
	if !equalPaths(report.Expected, report.Observed) || report.Coordinates.UnclassifiedPaths != 0 {
		report.Reason = "WRITE_SET_NOT_EXACT"
		return seal(report)
	}
	report.Decision, report.Reason, report.Exact = "PASS", "AUTHORIZED_WRITE_SET_EXACT", true
	return seal(report)
}

func unionPaths(left, right map[string]bool) map[string]bool {
	result := make(map[string]bool, len(left)+len(right))
	for value := range left {
		result[value] = true
	}
	for value := range right {
		result[value] = true
	}
	return result
}
