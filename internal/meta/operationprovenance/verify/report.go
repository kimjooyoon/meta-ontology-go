package verify

import "strings"

func producerImportCheck(source []byte) map[string]any {
	if len(source) == 0 {
		return map[string]any{"numerator": 0, "denominator": 1, "status": "UNKNOWN"}
	}
	if containsProducerImport(source) {
		return map[string]any{"numerator": 0, "denominator": 1, "status": "FAIL"}
	}
	return map[string]any{"numerator": 1, "denominator": 1, "status": "PASS"}
}

func containsProducerImport(source []byte) bool {
	return strings.Contains(string(source), "internal/meta/operationprovenance\"")
}

func verifiedReport(actual receipt, metrics []cMetric, imports map[string]any) (map[string]any, error) {
	result := map[string]any{
		"schema": "gooo/meta-operation-provenance-verification/v2", "status": "VERIFIED", "conformance_decision": "VERIFIED", "subject_resolution": "EXACT",
		"source_digest": actual.Source, "canonical_semantic_digest": actual.Semantic, "receipt_digest": actual.Digest,
		"scenario_count": len(actual.Scenarios), "metric_count": len(metrics) * len(actual.Scenarios),
		"fail_closed_count": 0, "direct_unknowns": 0, "dependency_blocks": 0,
		"transition_counts": map[string]int{}, "source_reconstruction": actual.Reconstruction, "producer_import": imports,
	}
	transitions := result["transition_counts"].(map[string]int)
	for _, scenario := range actual.Scenarios {
		for _, metric := range scenario.Metrics {
			switch metric.Decision {
			case "FAIL_CLOSED":
				result["fail_closed_count"] = result["fail_closed_count"].(int) + 1
			case "UNKNOWN":
				if metric.Issue != nil && metric.Issue.Cause == "DEPENDENCY_BLOCK" {
					result["dependency_blocks"] = result["dependency_blocks"].(int) + 1
				} else {
					result["direct_unknowns"] = result["direct_unknowns"].(int) + 1
				}
			}
			transitions[metric.Transition.Transition]++
		}
	}
	digest, err := digestJSON(result)
	result["digest"] = digest
	return result, err
}

func unknownReport(imports map[string]any) map[string]any {
	result := map[string]any{
		"schema": "gooo/meta-operation-provenance-verification/v2", "status": "UNKNOWN", "conformance_decision": "UNKNOWN", "subject_resolution": "LOWER_RESOLUTION",
		"scenario_count": 0, "metric_count": 0, "fail_closed_count": 0, "direct_unknowns": 0, "dependency_blocks": 0,
		"transition_counts": map[string]int{}, "source_reconstruction": sourceReconstruction{}, "producer_import": imports,
		"issue": map[string]any{"stage": "CONSUMER", "step": "parse-source", "reason": "REQUIRED_RAW_SOURCE_MISSING", "cause": "DIRECT_CAUSE"},
	}
	result["digest"], _ = digestJSON(result)
	return result
}
