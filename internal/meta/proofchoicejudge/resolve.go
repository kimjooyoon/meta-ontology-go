package proofchoicejudge

type derived struct {
	Items        []item
	Evidence     []evidence
	Compositions []composition
	Transitions  []transition
	Summary      summary
	Decision     string
	Reason       string
	Resolution   string
}

func resolve(values []value, source lowered, baseline []byte) derived {
	values, failure := splitValues(values)
	if failure != "" {
		return derived{Summary: baseSummary(), Decision: "FAIL_CLOSED", Reason: failure, Resolution: "LOWER_RESOLUTION"}
	}
	bundle := buildEvidence(values, source, baseline)
	result := derived{Decision: "PASS", Reason: "PROOF_VALUES_RESOLVED", Resolution: "EXACT", Evidence: bundle.All, Compositions: []composition{}}
	results := map[string]routeResult{}
	for _, current := range values {
		if current.Kind == "composition" {
			composition, route := composeValue(current, results)
			result.Compositions = append(result.Compositions, composition)
			results[current.ID] = route
			result.Items = append(result.Items, itemFor(current, route, bundle))
			result.Transitions = append(result.Transitions, transitionFor(len(result.Transitions)+1, current, route))
			if route.FailClosed {
				result.Decision, result.Reason, result.Resolution = "FAIL_CLOSED", route.Reason, "EXACT"
				break
			}
			continue
		}
		route := selectRoute(current, bundle)
		if validation := validateValue(current); validation != "" {
			route = routeResult{Resolution: "LOWER_RESOLUTION", ObservationState: "UNKNOWN", Reason: validation}
		}
		results[current.ID] = route
		result.Items = append(result.Items, itemFor(current, route, bundle))
		if current.Kind == "claim" {
			result.Transitions = append(result.Transitions, transitionFor(len(result.Transitions)+1, current, route))
		}
		if route.FailClosed {
			result.Decision, result.Reason, result.Resolution = "FAIL_CLOSED", route.Reason, "EXACT"
			break
		}
		if route.Resolution == "LOWER_RESOLUTION" && result.Resolution == "EXACT" {
			result.Resolution, result.Reason = "LOWER_RESOLUTION", route.Reason
		}
	}
	result.Summary = summarize(result.Items, result.Transitions, bundle)
	return result
}
