package integrationprogress

func summarize(results []pullEvaluation) (Summary, []Cell) {
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
	summary.QueueShareBasisPoints = basisPoints64(summary.QueueSecondsTotal,
		summary.QueueSecondsTotal+summary.ExecutionSecondsTotal)
	return summary, cells
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
		summary.QueueSecondsTotal += result.QueueSeconds
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

func basisPoints(numerator, denominator int) int {
	return basisPoints64(int64(numerator), int64(denominator))
}

func basisPoints64(numerator, denominator int64) int {
	if denominator == 0 {
		return 0
	}
	return int(numerator * 10000 / denominator)
}
