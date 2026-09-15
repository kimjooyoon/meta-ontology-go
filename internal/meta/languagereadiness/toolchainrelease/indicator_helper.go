package toolchainrelease

func indicator(id, class, proof string, value, target int, relation string) Indicator {
	satisfied := value >= target
	if relation == "less_or_equal" {
		satisfied = value <= target
	}
	return Indicator{
		MetricID: id, Class: class, ProofChoice: proof,
		Producer: metricProducer, Consumer: metricConsumer,
		MetaOperation: MetaOperation, Resolution: ResolutionExact,
		Value: value, Target: target, Relation: relation, Satisfied: satisfied,
	}
}

func buildIndicators(summary Summary) []Indicator {
	result := outcomeIndicators(summary)
	result = append(result, driverIndicators(summary)...)
	return append(result, guardrailIndicators(summary)...)
}
