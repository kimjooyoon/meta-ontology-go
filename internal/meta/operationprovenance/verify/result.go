package verify

func resultFor(metric cMetric, observations []relationObservation, forced string, forcedIssue *issue, sourceDigest, semanticDigest string, f cFixture) metricResult {
	result := metricResult{ID: metric.id, Family: metric.family, Claim: metric.claim, Denominator: 4, Proposition: proposition(metric.id, semanticDigest), SourceResolution: "EXACT", Relations: append([]relationObservation(nil), observations...)}
	for _, observation := range result.Relations {
		if observation.RelationStatus != "PASS" {
			continue
		}
		result.Numerator++
		switch observation.Relation {
		case "PRODUCES":
			result.Lineage.Producer = observation.ObservedEndpoint
		case "CONSUMES":
			result.Lineage.Consumer = observation.ObservedEndpoint
		case "OPERATES":
			result.Lineage.Operation = observation.ObservedEndpoint
		case "EVIDENCED_BY":
			result.Lineage.Evidence = observation.ObservedEndpoint
		}
	}
	result.LineageResolution = "EXACT"
	if result.Numerator != result.Denominator {
		result.LineageResolution = "LOWER_RESOLUTION"
	}
	result.Decision, result.Issue = decide(result, forced, forcedIssue)
	result.EvaluationState = "EVALUATED"
	result.Transition = transitionFor(result, sourceDigest, semanticDigest, f)
	return result
}

func decide(result metricResult, forced string, forcedIssue *issue) (string, *issue) {
	if forced != "" {
		return forced, forcedIssue
	}
	if result.Numerator == result.Denominator {
		return "PASS", nil
	}
	for _, observation := range result.Relations {
		if observation.RelationStatus != "PASS" {
			return "UNKNOWN", &issue{Stage: observation.Stage, Step: observation.Step, Reason: observation.Reason, Cause: observation.Cause}
		}
	}
	return "UNKNOWN", &issue{Stage: "LINEAGE", Step: "evaluate-relations", Reason: "RELATION_OBSERVATION_INCOMPLETE", Cause: "DIRECT_CAUSE"}
}

func proposition(metricID, semanticDigest string) string {
	return "metric " + metricID + " has complete executable lineage for this subject digest " + semanticDigest
}
