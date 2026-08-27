package integrationprogress

func buildIndicators(summary Summary) []Indicator {
	return []Indicator{
		targetIndicator("gooo.metric.integration-progress.cells-closed.v1", "OUTCOME", "foundation", int64(summary.ClosedCells), int64(summary.CellsTotal), "cells", equalityStatus(summary.ClosedCells, summary.CellsTotal)),
		targetIndicator("gooo.metric.integration-progress.merges.v1", "OUTCOME", "foundation", int64(summary.MergedPullRequests), int64(summary.PullRequestsTotal), "pull_requests", equalityStatus(summary.MergedPullRequests, summary.PullRequestsTotal)),
		targetIndicator("gooo.metric.integration-progress.evidence-reachable.v1", "DRIVER", "coherence", int64(summary.EvidenceReachable), int64(summary.PullRequestsTotal), "pull_requests", equalityStatus(summary.EvidenceReachable, summary.PullRequestsTotal)),
		targetIndicator("gooo.metric.integration-progress.evidenced-merges.v1", "OUTCOME", "coherence", int64(summary.EvidencedMerges), int64(summary.PullRequestsTotal), "pull_requests", equalityStatus(summary.EvidencedMerges, summary.PullRequestsTotal)),
		targetIndicator("gooo.metric.integration-progress.unknown-cells.v1", "GUARDRAIL", "foundation", int64(summary.UnknownCells), 0, "cells", zeroStatus(summary.UnknownCells, StateUnknown)),
		targetIndicator("gooo.metric.integration-progress.refuted-cells.v1", "GUARDRAIL", "coherence", int64(summary.RefutedCells), 0, "cells", zeroStatus(summary.RefutedCells, StateRefuted)),
		observedIndicator("gooo.metric.integration-progress.queue-seconds-total.v1", "DRIVER", "coherence", summary.QueueSecondsTotal, "seconds"),
		observedIndicator("gooo.metric.integration-progress.execution-seconds-total.v1", "DRIVER", "coherence", summary.ExecutionSecondsTotal, "seconds"),
		observedIndicator("gooo.metric.integration-progress.queue-share-bps.v1", "DRIVER", "coherence", int64(summary.QueueShareBasisPoints), "basis_points"),
		observedIndicator("gooo.metric.integration-progress.evidence-latency-seconds-total.v1", "DRIVER", "coherence", summary.EvidenceLatencySecondsTotal, "seconds"),
		observedIndicator("gooo.metric.integration-progress.merge-after-evidence-seconds-total.v1", "DRIVER", "coherence", summary.MergeAfterEvidenceSecondsTotal, "seconds"),
		targetIndicator("gooo.metric.integration-progress.repository-writes.v1", "GUARDRAIL", "regression", 0, 0, "writes", "SATISFIED"),
	}
}

func targetIndicator(id, class, proof string, value, target int64, unit, status string) Indicator {
	return Indicator{MetricID: id, Class: class, ProofChoice: proof, Value: value, Target: &target,
		Unit: unit, Relation: "EQUAL", Status: status, Producer: "integrationprogress.Evaluate",
		Consumer: "integration-progress-scorecard", MetaOperation: MetaOperation}
}

func observedIndicator(id, class, proof string, value int64, unit string) Indicator {
	return Indicator{MetricID: id, Class: class, ProofChoice: proof, Value: value, Unit: unit,
		Relation: "OBSERVE", Status: "OBSERVED", Producer: "integrationprogress.Evaluate",
		Consumer: "integration-progress-scorecard", MetaOperation: MetaOperation}
}

func equalityStatus(value, target int) string {
	if value == target {
		return "SATISFIED"
	}
	return "OPEN"
}

func zeroStatus(value int, failure string) string {
	if value == 0 {
		return "SATISFIED"
	}
	return failure
}
