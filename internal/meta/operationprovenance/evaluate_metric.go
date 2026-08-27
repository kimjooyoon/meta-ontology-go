package operationprovenance

func metricResult(metric metricSpec, edgeCounts map[string]int, forcedDecision string, forcedIssue *Issue, sourceDigest, semanticDigest string, f fixture) MetricResult {
	result := MetricResult{ID: metric.ID, Family: metric.Family, Claim: metric.PriorClaim, Denominator: relationDenominator, SubjectResolution: "EXACT", EvaluationState: "EVALUATED"}
	for _, link := range relations(metric) {
		if edgeCounts[link.From+"\x00"+link.To+"\x00"+link.Kind] != 1 {
			continue
		}
		result.Numerator++
		switch link.Kind {
		case "PRODUCES":
			result.Lineage.Producer = link.From
		case "CONSUMES":
			result.Lineage.Consumer = link.To
		case "OPERATES":
			result.Lineage.MetaOperation = link.From
		case "EVIDENCED_BY":
			result.Lineage.EvidencePath = link.To
		}
	}
	result.Decision, result.Issue = decideMetric(result, forcedDecision, forcedIssue)
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
	if result.Lineage.Consumer == "" {
		return "FAIL_CLOSED", &Issue{Stage: "LINEAGE", Step: "connect-consumer", Reason: "REQUIRED_CONSUMER_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	}
	if result.Lineage.Producer == "" {
		return "UNKNOWN", &Issue{Stage: "LINEAGE", Step: "connect-producer", Reason: "REQUIRED_PRODUCER_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	}
	if result.Lineage.MetaOperation == "" {
		return "UNKNOWN", &Issue{Stage: "LINEAGE", Step: "connect-meta-operation", Reason: "REQUIRED_META_OPERATION_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	}
	return "UNKNOWN", &Issue{Stage: "LINEAGE", Step: "connect-evidence-path", Reason: "REQUIRED_EVIDENCE_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
}
