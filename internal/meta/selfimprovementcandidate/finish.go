package selfimprovementcandidate

var reportNonClaims = []string{
	"candidate quality",
	"candidate feasibility",
	"value-level computation",
	"value-level witness existence",
	"compiler semantics changed",
	"execution or repository mutation",
	"promotion or automatic adoption",
}

func finish(report Report, success bool) Report {
	satisfied, eligible, target := 0, 0, 0
	if success {
		satisfied, eligible, target = indicatorTotal, 1, 1
	}
	report.Summary = Summary{Coordinates: coordinate(satisfied, indicatorTotal),
		SourceCoordinates: coordinate(satisfied, indicatorTotal), EligibleGaps: eligible,
		CandidateCount: len(report.Candidates), AchievedDelta: 0, TargetDelta: target, Unknowns: 0}
	report.Indicators = buildIndicators(success)
	report.Views = buildViews(report.Indicators)
	report.Proofs = buildProofs(report, success)
	report.NotClaimed = append([]string{}, reportNonClaims...)
	report.Digest = reportDigest(report)
	return report
}

func coordinate(satisfied, total int) Coordinate {
	basisPoints := 0
	if total > 0 {
		basisPoints = satisfied * 10_000 / total
	}
	return Coordinate{Satisfied: satisfied, Total: total, BasisPoints: basisPoints}
}

func coordinateEquals(value Coordinate, satisfied, total int) bool {
	return value == coordinate(satisfied, total)
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestJSON(report)
}
