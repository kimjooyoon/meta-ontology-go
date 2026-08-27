package integrationprogress

func Evaluate(observation Observation, replayVerified bool) Report {
	index, conflicts := indexObservation(observation)
	if !observationHeaderValid(observation) {
		conflicts++
	}
	results := make([]pullEvaluation, 0, len(PullNumbers()))
	for _, number := range PullNumbers() {
		pull, exists := index[number]
		if !exists {
			results = append(results, missingPull(number))
			continue
		}
		results = append(results, evaluatePull(pull))
	}
	summary, cells := summarize(results)
	summary.DenominatorConflicts = conflicts
	decision, reason, resolution := decide(summary)
	report := Report{
		Schema: ReportSchema, Repository: Repository, ObserverHeadSHA: observation.ObserverHeadSHA,
		CohortID: CohortID, Decision: decision, Reason: reason, Resolution: resolution,
		ObservationDigest: digestJSON(observation), MetaProgramDigest: digestBytes(RenderProgram()),
		Summary: summary, Cells: cells, RepositoryWrites: 0, PromotionAuthorized: false,
	}
	report.Indicators = buildIndicators(summary)
	report.Proofs = buildProofs(summary, replayVerified)
	return seal(report)
}

func decide(summary Summary) (string, string, string) {
	if summary.DenominatorConflicts > 0 || summary.RefutedCells > 0 {
		return "FAIL_CLOSED", "INTEGRATION_PROGRESS_CONTRADICTION", "INVARIANT_ONLY"
	}
	if summary.UnknownCells > 0 {
		return "LOWER_RESOLUTION", "INTEGRATION_PROGRESS_EVIDENCE_UNKNOWN", "CELL"
	}
	if summary.ClosedCells == summary.CellsTotal {
		return "COMPLETE", "INTEGRATION_PROGRESS_COMPLETE", "EXACT"
	}
	return "PROGRESS_OBSERVED", "INTEGRATION_PROGRESS_OPEN", "EXACT"
}
