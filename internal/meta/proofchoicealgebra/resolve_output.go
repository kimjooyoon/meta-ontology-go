package proofchoicealgebra

func itemFor(value Value, route routeResult, bundle evidenceBundle) Item {
	item := Item{Kind: value.Kind, ID: value.ID, Statement: value.Statement, PriorState: value.PriorState, Choice: string(route.Route), Resolution: route.Resolution, ObservationState: route.ObservationState, Observations: route.Observations, EvidenceDigest: route.EvidenceDigest, Provenance: route.Provenance}
	if value.Kind == MetricKind {
		item.Denominator = FixedDenom
		for _, evidence := range evidenceFor(value, bundle) {
			for _, slot := range evidence.ObservationSlots {
				if slot.Observed {
					item.Numerator++
				}
			}
			if len(evidence.ObservationSlots) > 0 {
				break
			}
		}
	}
	return item
}
