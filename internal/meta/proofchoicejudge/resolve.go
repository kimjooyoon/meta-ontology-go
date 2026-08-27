package proofchoicejudge

type derived struct {
	Items       []item
	Transitions []transition
	Summary     summary
	Decision    string
	Reason      string
	Resolution  string
}

func resolve(values []value) derived {
	observations, subjects, failure, conflict := splitValues(values)
	if failure != "" {
		result := derived{Summary: baseSummary(), Decision: "FAIL_CLOSED", Reason: failure, Resolution: "FAIL_CLOSED"}
		if conflict.Kind == "claim" {
			result.Transitions = []transition{transitionFor(1, conflict, routeResult{Resolution: "FAIL_CLOSED", Reason: failure, Observations: conflict.Observations}, observations)}
		}
		return result
	}
	result := derived{Decision: "PASS", Reason: "PROOF_VALUES_RESOLVED", Resolution: "EXACT"}
	for _, subject := range subjects {
		if subject.Kind == "metric" && len(subject.Slots) != 3 {
			result.Decision, result.Reason, result.Resolution = "FAIL_CLOSED", "FIXED_DENOMINATOR_MISMATCH", "FAIL_CLOSED"
			result.Summary = summarize(result.Items, result.Transitions, len(observations))
			return result
		}
		route := selectRoute(subject, observations)
		if failure := validateSubject(subject, observations); failure != "" && route.Resolution != "FAIL_CLOSED" {
			route = routeResult{Resolution: "UNKNOWN", Reason: failure, Observations: subject.Observations}
		}
		result.Items = append(result.Items, itemFor(subject, route, observations))
		if subject.Kind == "claim" {
			result.Transitions = append(result.Transitions, transitionFor(len(result.Transitions)+1, subject, route, observations))
		}
		if route.Resolution == "FAIL_CLOSED" {
			result.Decision, result.Reason, result.Resolution = "FAIL_CLOSED", route.Reason, "FAIL_CLOSED"
			break
		}
		if route.Resolution == "UNKNOWN" && result.Resolution == "EXACT" {
			result.Resolution, result.Reason = "UNKNOWN", route.Reason
		}
		if route.Resolution == "LOWER_RESOLUTION" && result.Resolution == "EXACT" {
			result.Resolution, result.Reason = "LOWER_RESOLUTION", route.Reason
		}
	}
	result.Summary = summarize(result.Items, result.Transitions, len(observations))
	return result
}
