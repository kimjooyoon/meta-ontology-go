package verify

func resultFor(metric cMetric, edges map[string]int, forced string, forcedIssue *issue, sourceDigest, semanticDigest string, f cFixture) metricResult {
	result := metricResult{ID: metric.id, Family: metric.family, Claim: metric.claim, Denominator: 4, Resolution: "EXACT", EvaluationState: "EVALUATED"}
	for _, link := range links(metric) {
		if edges[link.from+"\x00"+link.to+"\x00"+link.kind] != 1 {
			continue
		}
		result.Numerator++
		switch link.kind {
		case "PRODUCES":
			result.Lineage.Producer = link.from
		case "CONSUMES":
			result.Lineage.Consumer = link.to
		case "OPERATES":
			result.Lineage.Operation = link.from
		case "EVIDENCED_BY":
			result.Lineage.Evidence = link.to
		}
	}
	result.Decision, result.Issue = decide(result, forced, forcedIssue)
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
	if result.Lineage.Consumer == "" {
		return "FAIL_CLOSED", &issue{Stage: "LINEAGE", Step: "connect-consumer", Reason: "REQUIRED_CONSUMER_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	}
	if result.Lineage.Producer == "" {
		return "UNKNOWN", &issue{Stage: "LINEAGE", Step: "connect-producer", Reason: "REQUIRED_PRODUCER_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	}
	if result.Lineage.Operation == "" {
		return "UNKNOWN", &issue{Stage: "LINEAGE", Step: "connect-meta-operation", Reason: "REQUIRED_META_OPERATION_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	}
	return "UNKNOWN", &issue{Stage: "LINEAGE", Step: "connect-evidence-path", Reason: "REQUIRED_EVIDENCE_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
}
