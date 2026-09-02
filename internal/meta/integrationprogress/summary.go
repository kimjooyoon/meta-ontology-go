package integrationprogress

func summarize(results []pullEvaluation, queue QueueSnapshot) (Summary, []Cell) {
	summary := Summary{PullRequestsTotal: len(PullNumbers()), CellsTotal: CellDenominator()}
	cells := make([]Cell, 0, summary.CellsTotal)
	for _, result := range results {
		cells = append(cells, result.Cells...)
		accumulateStates(&summary, result.Cells)
		accumulateTimings(&summary, result)
	}
	summary.ProgressBasisPoints = basisPoints(summary.ClosedCells, summary.CellsTotal)
	summary.MergeBasisPoints = basisPoints(summary.MergedPullRequests, summary.PullRequestsTotal)
	summary.EvidenceBasisPoints = basisPoints(summary.EvidenceReachable, summary.PullRequestsTotal)
	applyQueueSnapshot(&summary, queue)
	return summary, cells
}

func applyQueueSnapshot(summary *Summary, queue QueueSnapshot) {
	if queue.ObservationStatus != "OBSERVED" || queue.QueuedRuns < 0 || queue.InProgressRuns < 0 {
		summary.QueueObservationUnknown = 1
		return
	}
	summary.QueuedRunsSnapshot = queue.QueuedRuns
	summary.InProgressRunsSnapshot = queue.InProgressRuns
	summary.QueuePressureBasisPoints = basisPoints(queue.QueuedRuns, queue.QueuedRuns+queue.InProgressRuns)
}

func accumulateStates(summary *Summary, cells []Cell) {
	for _, value := range cells {
		switch value.State {
		case StateClosed:
			summary.ClosedCells++
		case StateOpen:
			summary.OpenCells++
		case StateUnknown:
			summary.UnknownCells++
		case StateRefuted:
			summary.RefutedCells++
		}
	}
	if cells[0].State == StateClosed {
		summary.PullRequestsObserved++
	}
	if cells[1].State == StateClosed {
		summary.TerminalRuns++
	}
	if cells[2].State == StateClosed {
		summary.EvidenceReachable++
	}
	if cells[3].State == StateClosed {
		summary.MergedPullRequests++
	}
	if cells[4].State == StateClosed {
		summary.EvidencedMerges++
	}
}

func accumulateTimings(summary *Summary, result pullEvaluation) {
	if result.TimingSample {
		summary.TimingSamples++
		summary.RunStartDelaySecondsTotal += result.QueueSeconds
		summary.ExecutionSecondsTotal += result.ExecutionSeconds
	}
	if result.EvidenceLatencySample {
		summary.EvidenceLatencySamples++
		summary.EvidenceLatencySecondsTotal += result.EvidenceLatencySeconds
	}
	if result.MergeDelaySample {
		summary.MergeDelaySamples++
		summary.MergeAfterEvidenceSecondsTotal += result.MergeAfterEvidenceSeconds
	}
}
