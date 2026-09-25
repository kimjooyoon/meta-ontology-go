package ciplanusecase

import "reflect"

func summarize(input Input) Summary {
	summary := Summary{GoooFiles: input.Source.GoooFiles, GoFiles: input.Source.GoFiles, GoooLines: input.Source.GoooLines, GoLines: input.Source.GoLines}
	if input.GeneratedReplay {
		summary.GeneratedReplays = 1
	}
	for _, spec := range input.Contract.Cases {
		report, ok := input.Reports[spec.ID]
		if !ok {
			continue
		}
		applyInvocationSummary(&summary, report)
		if replay, exists := input.Replays[spec.ID]; exists && replay.ReportDigest == report.ReportDigest {
			summary.DeterministicReplays++
		}
		if golden, exists := input.Goldens[spec.ID]; exists && reflect.DeepEqual(ProjectGolden(report), golden) {
			summary.GoldenPlans++
		}
	}
	applyProfileSummary(&summary, input)
	return summary
}
