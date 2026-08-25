package main

func sourceIssue(expectedSHA string, density densityReport, extraction extractionReport) (string, int) {
	densityExact := density.Schema == densitySchema && density.SourceSHA == expectedSHA
	extractionExact := extraction.Schema == extractionSchema && extraction.SourceSHA == expectedSHA
	receipts := boolInt(densityExact) + boolInt(extractionExact)
	if !densityExact {
		return "DENSITY_RECEIPT_UNKNOWN", receipts
	}
	if !extractionExact {
		return "EXTRACTION_RECEIPT_UNKNOWN", receipts
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
