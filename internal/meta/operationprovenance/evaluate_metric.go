package operationprovenance

func metricResult(metric metricSpec, observations []RelationObservation, forced string, issue *Issue, sourceDigest, semanticDigest string, f fixture) MetricResult {
	result := MetricResult{ID: metric.ID, Family: metric.Family, Claim: metric.PriorClaim, Denominator: relationDenominator, Proposition: proposition(metric.ID, semanticDigest), SourceResolution: "EXACT", Relations: append([]RelationObservation(nil), observations...)}
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
			result.Lineage.MetaOperation = observation.ObservedEndpoint
		case "EVIDENCED_BY":
			result.Lineage.EvidencePath = observation.ObservedEndpoint
		}
	}
	result.LineageResolution = "EXACT"
	if result.Numerator != result.Denominator {
		result.LineageResolution = "LOWER_RESOLUTION"
	}
	result.Decision, result.Issue = decideMetric(result, forced, issue)
	result.EvaluationState = "EVALUATED"
	result.Transition = makeTransition(result, sourceDigest, semanticDigest, f)
	return result
}

func decideMetric(result MetricResult, forced string, issue *Issue) (string, *Issue) {
	if forced != "" {
		return forced, issue
	}
	if result.Numerator == result.Denominator {
		return "PASS", nil
	}
	for _, observation := range result.Relations {
		if observation.RelationStatus != "PASS" {
			return "UNKNOWN", &Issue{Stage: observation.Stage, Step: observation.Step, Reason: observation.Reason, Cause: observation.Cause}
		}
	}
	return "UNKNOWN", &Issue{Stage: "LINEAGE", Step: "evaluate-relations", Reason: "RELATION_OBSERVATION_INCOMPLETE", Cause: "DIRECT_CAUSE"}
}

func proposition(metricID, semanticDigest string) string {
	return "metric " + metricID + " has complete executable lineage for this subject digest " + semanticDigest
}
