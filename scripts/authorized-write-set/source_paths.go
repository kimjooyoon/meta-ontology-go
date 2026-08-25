package main

func sourceIssue(expectedSHA string, density densityReport, extraction extractionReport, split splitReport) (string, int) {
	densityExact := density.Schema == densitySchema && density.SourceSHA == expectedSHA
	extractionExact := extraction.Schema == extractionSchema && extraction.SourceSHA == expectedSHA
	splitExact := split.Schema == splitSchema && split.SourceSHA == expectedSHA
	receipts := boolInt(densityExact) + boolInt(extractionExact) + boolInt(splitExact)
	if !densityExact {
		return "DENSITY_RECEIPT_UNKNOWN", receipts
	}
	if !extractionExact {
		return "EXTRACTION_RECEIPT_UNKNOWN", receipts
	}
	if !splitExact {
		return "SPLIT_RECEIPT_UNKNOWN", receipts
	}
	return "", receipts
}

func densityPaths(report densityReport) (map[string]bool, bool) {
	result := map[string]bool{}
	for _, subject := range report.Subjects {
		if !canonicalPath(subject.Logical) {
			return nil, false
		}
		switch subject.Status {
		case "applied":
			result[subject.Logical] = true
		case "blocked":
		default:
			return nil, false
		}
	}
	return result, true
}

func extractionPaths(report extractionReport) (map[string]bool, map[string]bool, bool) {
	values, created := make([]string, 0), make([]string, 0)
	for _, subject := range report.Subjects {
		values = append(values, subject.Files...)
		created = append(created, subject.Created...)
	}
	for _, value := range report.Unhandled {
		if !canonicalPath(value) {
			return nil, nil, false
		}
	}
	paths, pathsExact := pathSet(values)
	createdPaths, createdExact := pathSet(created)
	if !pathsExact || !createdExact {
		return nil, nil, false
	}
	for value := range createdPaths {
		if !paths[value] {
			return nil, nil, false
		}
	}
	return paths, createdPaths, true
}

func splitPaths(report splitReport) (map[string]bool, map[string]bool, bool) {
	decisionExact := report.Decision == "PASS" || report.Decision == "FIXED_POINT"
	coordinates := report.Coordinates
	if !decisionExact || report.Resolution != "EXACT" || !report.Exact || coordinates.Unknowns != 0 ||
		coordinates.Selected != coordinates.Applied || coordinates.Applied != len(report.Subjects) {
		return nil, nil, false
	}
	paths, created, exact := extractionPaths(extractionReport{Subjects: report.Subjects, Unhandled: report.Unhandled})
	if !exact || len(paths) != coordinates.Changed || len(created) != coordinates.Created {
		return nil, nil, false
	}
	return paths, created, true
}
