package toolchainlsp

func indicator(name, class, proof string, value, target int, relation, resolution string) Indicator {
	satisfied := value == target
	if relation == "greater_or_equal" {
		satisfied = value >= target
	}
	if relation == "less_or_equal" {
		satisfied = value <= target
	}
	return Indicator{
		MetricID: metricPrefix + name, Class: class, ProofChoice: proof,
		Producer: "toolchainlsp.Evaluate", Consumer: "self-improvement-cycle",
		MetaOperation: MetaOperation, Resolution: resolution, Value: value,
		Target: target, Relation: relation, Satisfied: satisfied,
	}
}

func buildIndicators(summary Summary, resolution string) []Indicator {
	result := outcomeIndicators(summary, resolution)
	result = append(result, driverIndicators(summary, resolution)...)
	return append(result, guardrailIndicators(summary, resolution)...)
}
