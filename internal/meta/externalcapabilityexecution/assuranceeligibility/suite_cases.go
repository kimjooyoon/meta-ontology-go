package assuranceeligibility

func CaseInput(base Input, id string) (Input, bool) {
	input := base.clone()
	switch id {
	case "exact":
	case "missing-assurance":
		input.Payloads[AssuranceName] = nil
	case "missing-parent-report":
		input.Payloads[ParentReportName] = nil
	case "missing-parent-observation":
		input.Payloads[ParentObservationName] = nil
	case "missing-parent-suite":
		input.Payloads[ParentSuiteName] = nil
	case "missing-capability-report":
		input.Payloads[CapabilityReportName] = nil
	case "missing-capability-observation":
		input.Payloads[CapabilityObservationName] = nil
	case "missing-capability-suite":
		input.Payloads[CapabilitySuiteName] = nil
	case "unknown-assurance-state":
		rewrite(input, AssuranceName, func(value map[string]any) {
			for _, item := range value["obligations"].([]any) {
				obligation := item.(map[string]any)
				if obligation["metric_id"] == MetricID {
					obligation["status"] = "UNKNOWN"
				}
			}
		})
	case "unknown-parent-decision":
		rewrite(input, ParentReportName, set("decision", "UNKNOWN"))
	case "unknown-capability-decision":
		rewrite(input, CapabilityReportName, set("decision", "UNKNOWN"))
	case "assurance-digest-mismatch":
		input.Payloads[AssuranceName] = append(input.Payloads[AssuranceName], '\n')
	case "parent-false-fixed":
		rewrite(input, ParentReportName, set("decision", "FIXED_POINT"))
	case "parent-count-mismatch":
		rewrite(input, ParentReportName, set("completed", 5))
	case "capability-count-mismatch":
		rewrite(input, CapabilityReportName, set("completed", 9))
	case "capability-suite-mismatch":
		rewrite(input, CapabilitySuiteName, set("passed", 14))
	case "reference-mismatch":
		rewrite(input, CapabilityObservationName, func(value map[string]any) {
			value["reference"].(map[string]any)["commit_sha"] = "0000000000000000000000000000000000000000"
		})
	case "observed-write":
		rewrite(input, CapabilityReportName, set("repository_writes", 1))
	case "official-mutation":
		rewrite(input, CapabilityReportName, set("official_mutation_count", 1))
	case "promotion-observed":
		rewrite(input, CapabilityReportName, set("promotion_count", 1))
	default:
		return Input{}, false
	}
	return input, true
}
