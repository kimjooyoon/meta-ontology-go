package integrationprogress

func buildProofs(summary Summary, replayVerified bool) []Proof {
	coherence := "SATISFIED"
	if summary.RefutedCells > 0 || summary.DenominatorConflicts > 0 {
		coherence = StateRefuted
	} else if summary.UnknownCells > 0 || summary.QueueObservationUnknown > 0 {
		coherence = StateUnknown
	}
	replay := StateRefuted
	if replayVerified {
		replay = "SATISFIED"
	}
	return []Proof{
		{Choice: "foundation", Status: "SATISFIED", Evidence: []string{CohortID, "30 pull requests", "150 progress cells"}},
		{Choice: "coherence", Status: coherence, Evidence: []string{"pull/head", "run/head", "artifact/head", "artifact-created-at<=merged-at"}},
		{Choice: "regression", Status: replay, Evidence: []string{"report byte replay", "generated Gooo byte replay"}},
	}
}
