package main

func sourceBinding(rows []sourceIndicator, route string) metaBinding {
	applicable, notApplicable := 0, 0
	for _, row := range rows {
		if row.Applicability == "APPLICABLE" {
			applicable++
		}
		if row.Applicability == "NOT_APPLICABLE" {
			notApplicable++
		}
	}
	return metaBinding{Kind: "SOURCE_INDICATORS", Operation: operationSet(rows), Route: route, IndicatorCount: len(rows), ApplicableIndicators: applicable, NotApplicableIndicators: notApplicable, IndicatorDigest: digestValues(rows)}
}

func derivedBinding(kind, operation, route string) metaBinding {
	return metaBinding{Kind: kind, Operation: operation, Route: route, IndicatorDigest: digestValues([]sourceIndicator{})}
}
