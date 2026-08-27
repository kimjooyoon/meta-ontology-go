package proofchoicealgebra

func resolve(values []Value, lowered lowered, baseline []byte) ([]Item, []Evidence, []Composition, []Transition, Summary, string, string, string) {
	values, failure := splitValues(values)
	if failure != "" {
		return nil, nil, nil, nil, baseSummary(), FailClosed, failure, Lower
	}
	bundle := buildEvidence(values, lowered, baseline)
	items := make([]Item, 0, len(values))
	transitions := []Transition{}
	compositions := []Composition{}
	results := map[string]routeResult{}
	decision, reason, resolution := Pass, "PROOF_VALUES_RESOLVED", Exact
	for _, value := range values {
		if value.Kind == CompositionKind {
			composition, route := composeValue(value, results)
			compositions = append(compositions, composition)
			results[value.ID] = route
			items = append(items, itemFor(value, route, bundle))
			transitions = append(transitions, transitionFor(len(transitions)+1, value, route))
			if route.FailClosed {
				return items, bundle.All, compositions, transitions, summarize(items, transitions, bundle), FailClosed, route.Reason, Exact
			}
			continue
		}
		route := selectRoute(value, bundle)
		if validation := validateValue(value); validation != "" {
			route = routeResult{Resolution: Lower, ObservationState: UnknownState, Reason: validation}
		}
		results[value.ID] = route
		items = append(items, itemFor(value, route, bundle))
		if value.Kind == ClaimKind {
			transitions = append(transitions, transitionFor(len(transitions)+1, value, route))
		}
		if route.FailClosed {
			return items, bundle.All, compositions, transitions, summarize(items, transitions, bundle), FailClosed, route.Reason, Exact
		}
		if route.Resolution == Lower && resolution == Exact {
			resolution, reason = Lower, route.Reason
		}
	}
	return items, bundle.All, compositions, transitions, summarize(items, transitions, bundle), decision, reason, resolution
}
