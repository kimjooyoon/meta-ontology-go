package integrationprogress

func evaluatePull(value PullObservation) pullEvaluation {
	stages := StageSpecs()
	pull := evaluatePullObservation(value, stages[0])
	run, timing := evaluateRun(value, stages[1], pull)
	artifact, artifactAt := evaluateArtifact(value, stages[2], run, timing)
	merge, mergedAt := evaluateMerge(value, stages[3], pull)
	link := evaluateLink(value.Number, stages[4], artifact, merge, artifactAt, mergedAt)
	result := pullEvaluation{Cells: []Cell{pull, run, artifact, merge, link}}
	if run.State == StateClosed {
		result.TimingSample = true
		result.QueueSeconds = int64(timing.Started.Sub(timing.Created).Seconds())
		result.ExecutionSeconds = int64(timing.Updated.Sub(timing.Started).Seconds())
	}
	if artifact.State == StateClosed {
		result.EvidenceLatencySample = true
		result.EvidenceLatencySeconds = int64(artifactAt.Sub(timing.Created).Seconds())
	}
	if link.State == StateClosed {
		result.MergeDelaySample = true
		result.MergeAfterEvidenceSeconds = int64(mergedAt.Sub(artifactAt).Seconds())
	}
	return result
}

func missingPull(number int) pullEvaluation {
	result := pullEvaluation{Cells: make([]Cell, 0, len(StageSpecs()))}
	for _, stage := range StageSpecs() {
		result.Cells = append(result.Cells, cell(number, stage, StateUnknown, "PULL_DENOMINATOR_MEMBER_MISSING", ""))
	}
	return result
}
