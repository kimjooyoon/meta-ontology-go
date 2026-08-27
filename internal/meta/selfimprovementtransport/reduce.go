package selfimprovementtransport

func reduce(report *Report) {
	report.Metrics.FixedObligationTotal = fixedObligationTotal
	var firstFalse, firstUnknown *Obligation
	for index := range report.Obligations {
		obligation := &report.Obligations[index]
		switch obligation.Status {
		case StatusVerified:
			report.Metrics.VerifiedTotal++
		case StatusUnknown:
			report.Metrics.UnknownTotal++
			report.OpenObligationIDs = append(report.OpenObligationIDs, obligation.ID)
			if firstUnknown == nil {
				firstUnknown = obligation
			}
		default:
			report.Metrics.FalseTotal++
			report.OpenObligationIDs = append(report.OpenObligationIDs, obligation.ID)
			if firstFalse == nil {
				firstFalse = obligation
			}
		}
	}
	report.Metrics.OpenTotal = report.Metrics.UnknownTotal + report.Metrics.FalseTotal
	report.Metrics.CoverageBasisPoints = report.Metrics.VerifiedTotal * 10000 / fixedObligationTotal
	switch {
	case firstFalse != nil:
		report.Decision, report.Resolution = DecisionFailClosed, ResolutionLower
		report.Reason, report.Coordinate = ReasonKnownMismatch, firstFalse.Coordinate
	case firstUnknown != nil:
		report.Decision, report.Resolution = DecisionObserved, ResolutionLower
		report.Reason, report.Coordinate = firstUnknown.Reason, firstUnknown.Coordinate
	default:
		report.Decision, report.Resolution = DecisionPass, ResolutionExact
		report.Reason = ReasonComplete
		report.Coordinate = Coordinate{Stage: "REDUCE", Step: "close-eht8"}
	}
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestJSON(report)
}
