package metarecognition

func baselineOutcome(state State, reason Reason, ids []string, b BaselineConfig, selected int) Outcome {
	if selected > b.FullCommands {
		selected = b.FullCommands
	}
	work := Work{Selected: selected, Full: b.FullCommands, ProvRecords: b.ProvRecords, ProvPaths: b.ProvPaths}
	work.Units = work.Selected
	return Outcome{State: state, Reason: reason, LocalizedIDs: sorted(ids), Work: work}
}
