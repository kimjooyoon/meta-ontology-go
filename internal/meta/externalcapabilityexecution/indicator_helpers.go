package externalcapabilityexecution

func knownParent(parent ParentReport) bool {
	decisionKnown := parent.Decision == DecisionFailClosed || parent.Decision == "EXECUTION_COMPATIBLE"
	resolutionKnown := parent.Resolution == ResolutionExact || parent.Resolution == ResolutionInvariant
	return decisionKnown && resolutionKnown
}

func indicator(suffix, class, proof, operation string, known, exact bool, value, target int) Indicator {
	metricStatus := StatusUnknown
	resolution := ResolutionUnknown
	if known {
		metricStatus = status(exact)
		resolution = ResolutionInvariant
		if exact {
			resolution = ResolutionExact
		}
	}
	relation := "GREATER_OR_EQUAL"
	if suffix == "writes" {
		relation = "LESS_OR_EQUAL"
	}
	return Indicator{
		MetricID: "gooo.metric.external-capability." + suffix + ".v1",
		Class: class, ProofChoice: proof,
		Producer: "externalcapabilityexecution.Evaluate",
		Consumer: "external-capability-go127-gate",
		MetaOperation: operation, Unit: "checks", Relation: relation,
		Resolution: resolution, Value: value, Target: target, Status: metricStatus,
	}
}
