package main

func reduce(expectedSHA string, density densityReport, extraction extractionReport,
	observed, untracked []string) evidence {
	report := evidence{Schema: evidenceSchema, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
		Reason: "WRITE_SET_EVIDENCE_UNKNOWN", Audience: "GOVERNOR", SourceSHA: expectedSHA,
		MetaOperation: metaOperation, Coordinates: coordinates{SourceReceiptsTotal: 2}}
	if issue, receipts := sourceIssue(expectedSHA, density, extraction); issue != "" {
		report.Reason, report.Coordinates.SourceReceipts = issue, receipts
		report.Coordinates.Unknowns = 2 - receipts
		return seal(report)
	}
	report.Coordinates.SourceReceipts = 2
	densitySet, densityExact := densityPaths(density)
	if !densityExact {
		report.Reason, report.Coordinates.SourceReceipts, report.Coordinates.Unknowns = "DENSITY_PATH_UNKNOWN", 1, 1
		return seal(report)
	}
	extractionSet, createdSet, extractionExact := extractionPaths(extraction)
	if !extractionExact {
		report.Reason, report.Coordinates.SourceReceipts, report.Coordinates.Unknowns = "EXTRACTION_PATH_UNKNOWN", 1, 1
		return seal(report)
	}
	observedSet, observedExact := pathSet(observed)
	untrackedSet, untrackedExact := pathSet(untracked)
	if !observedExact || !untrackedExact {
		report.Reason, report.Coordinates.Unknowns = "OBSERVED_WRITE_SET_UNKNOWN", 1
		return seal(report)
	}
	expectedSet := unionPaths(densitySet, extractionSet)
	report.Expected, report.Observed = sortedPaths(expectedSet), sortedPaths(observedSet)
	report.ExpectedCreated, report.ObservedCreated = sortedPaths(createdSet), sortedPaths(untrackedSet)
	report.Coordinates.DensityPaths, report.Coordinates.ExtractionPaths = len(densitySet), len(extractionSet)
	report.Coordinates.OverlapPaths = overlapCount(densitySet, extractionSet)
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
