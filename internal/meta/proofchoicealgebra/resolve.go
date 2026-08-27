package proofchoicealgebra

func resolve(values []Value) ([]Item, []Transition, Summary, string, string, string) {
	observations, subjects, failure := splitValues(values)
	summary := Summary{FixedDenominator: FixedDenom, ChoiceCounts: map[Route]int{Foundation: 0, Coherence: 0, Regression: 0}}
	if failure != "" {
		transitions := []Transition{}
		if len(subjects) == 1 && subjects[0].Kind == ClaimKind {
			transitions = append(transitions, transitionFor(1, subjects[0], routeResult{Resolution: FailClosed, Reason: failure, Observations: subjects[0].Observations}, observations))
		}
		return nil, transitions, summary, FailClosed, failure, FailClosed
	}
	items := make([]Item, 0, len(subjects))
	transitions := []Transition{}
	resolution := Exact
	reason := "PROOF_VALUES_RESOLVED"
	for _, value := range subjects {
		if value.Kind == MetricKind && len(value.Slots) != FixedDenom {
			return items, transitions, summarize(items, transitions, len(observations)), FailClosed, "FIXED_DENOMINATOR_MISMATCH", FailClosed
		}
		route := selectRoute(value, observations)
		if failure := validateDependencies(value, observations); failure != "" && route.Resolution != FailClosed {
			route = routeResult{Resolution: Unknown, Reason: failure, Observations: value.Observations}
		}
		item := itemFor(value, route, observations)
		items = append(items, item)
		if value.Kind == ClaimKind {
			transitions = append(transitions, transitionFor(len(transitions)+1, value, route, observations))
		}
		if route.Resolution == FailClosed {
			return items, transitions, summarize(items, transitions, len(observations)), FailClosed, route.Reason, FailClosed
		}
		if route.Resolution == Unknown && resolution == Exact {
			resolution, reason = Unknown, route.Reason
		}
		if route.Resolution == Lower && resolution == Exact {
			resolution, reason = Lower, route.Reason
		}
	}
	summary = summarize(items, transitions, len(observations))
	return items, transitions, summary, Pass, reason, resolution
}

func itemFor(value Value, route routeResult, observations map[string]Value) Item {
	item := Item{Kind: value.Kind, ID: value.ID, Statement: value.Statement, PriorState: value.PriorState, Choice: route.Route, Resolution: route.Resolution, Observations: route.Observations, EvidenceDigest: evidenceDigest(route.Observations, observations), Provenance: evidenceProvenance(route.Observations, observations)}
	if value.Kind == MetricKind {
		item.Denominator = len(value.Slots)
		for _, slot := range value.Slots {
			if slot.Observed {
				item.Numerator++
			}
		}
	}
	return item
}
