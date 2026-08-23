package toolchainrelease

func driverIndicators(s Summary) []Indicator {
	values := []int{
		s.CasesSatisfied, s.PlatformReceipts, s.OperatingSystems, s.Architectures,
		s.BinaryBuilds, s.ArchiveBuilds, s.NativeSmokes, s.BinaryReplays,
		s.ArchiveReplays, s.ChecksumEntries, s.ToolchainBindings, s.VCSBindings,
		s.ConceptBindings, s.CodeBindings, s.MetricBindings, s.UseCaseBindings,
	}
	targets := []int{CaseCount, 3, 3, 1, 6, 6, 3, 3, 3, 3, 3, 3, 1, 6, IndicatorCount, 3}
	proofs := []string{
		"COHERENCE", "FOUNDATION", "FOUNDATION", "FOUNDATION",
		"REGRESSION", "REGRESSION", "COHERENCE", "REGRESSION",
		"REGRESSION", "COHERENCE", "FOUNDATION", "FOUNDATION",
		"FOUNDATION", "FOUNDATION", "COHERENCE", "REGRESSION",
	}
	result := make([]Indicator, 0, DriverCount)
	for index, id := range driverMetricIDs {
		result = append(result, indicator(id, "DRIVER", proofs[index],
			values[index], targets[index], "greater_or_equal"))
	}
	return result
}
