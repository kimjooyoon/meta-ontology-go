package toolchainconformance

func metric(id, class, proof string, value, target int,
	relation string) Indicator {
	satisfied := value >= target
	if relation == "less_or_equal" {
		satisfied = value <= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "toolchainconformance.Evaluate",
		Consumer: "self-improvement-cycle",
		MetaOperation: ExpectedMetaOperation, Resolution: ResolutionExact,
		Value: value, Target: target, Relation: relation, Satisfied: satisfied}
}

func metricIDs() []string {
	indicators := buildIndicators(Summary{})
	ids := make([]string, len(indicators))
	for index := range indicators {
		ids[index] = indicators[index].MetricID
	}
	return ids
}
