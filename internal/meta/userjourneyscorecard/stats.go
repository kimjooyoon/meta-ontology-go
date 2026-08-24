package userjourneyscorecard

import "sort"

func summarizeJourney(definition JourneyDefinition, samples []Sample, contract Contract) (JourneyStats, bool) {
	stats := JourneyStats{ID: definition.ID, Operation: definition.Operation, Arguments: definition.Arguments, Samples: len(samples)}
	walls, memory := []int64{}, []int64{}
	complete := len(samples) == contract.SamplesPerJourney
	stdoutDigest, stderrDigest := "", ""
	sequences := map[int]bool{}
	for _, sample := range samples {
		if sample.Sequence < 1 || sample.Sequence > contract.SamplesPerJourney || sequences[sample.Sequence] {
			complete = false
		}
		sequences[sample.Sequence] = true
		if sample.ExitCode == 0 {
			stats.Successful++
		}
		if stdoutDigest == "" {
			stdoutDigest, stderrDigest = sample.StdoutDigest, sample.StderrDigest
		}
		if stdoutDigest != sample.StdoutDigest || stderrDigest != sample.StderrDigest {
			complete = false
		}
		walls, memory = append(walls, sample.WallMS), append(memory, sample.MaxRSSKiB)
		stats.StdoutMaxBytes = max(stats.StdoutMaxBytes, sample.StdoutBytes)
		stats.StderrMaxBytes = max(stats.StderrMaxBytes, sample.StderrBytes)
	}
	stats.OutputReplay = complete && len(samples) > 0
	if len(samples) > 0 {
		sort.Slice(walls, func(i, j int) bool { return walls[i] < walls[j] })
		sort.Slice(memory, func(i, j int) bool { return memory[i] < memory[j] })
		middle := len(samples) / 2
		stats.WallMinMS, stats.WallMedianMS, stats.WallMaxMS = walls[0], walls[middle], walls[len(walls)-1]
		stats.RSSMinKiB, stats.RSSMedianKiB, stats.RSSMaxKiB = memory[0], memory[middle], memory[len(memory)-1]
	}
	stats.EnvelopePassed = stats.Successful == contract.SamplesPerJourney && stats.OutputReplay &&
		stats.WallMaxMS <= contract.WallMSLimit && stats.RSSMaxKiB <= contract.MaxRSSKiBLimit
	return stats, complete
}

func max(left, right int64) int64 {
	if right > left {
		return right
	}
	return left
}
