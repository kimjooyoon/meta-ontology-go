package metarecognition

func productionOutcome(state State, reason Reason, ids []string, work Work) Outcome {
	work.Units = work.Selected
	return Outcome{State: state, Reason: reason, LocalizedIDs: sorted(ids), Work: work}
}
