package completenessdelta

func observed(d *DimensionObservation) bool { return d.Status == "PASS" || d.Status == "PROGRESS" }

func change(id string, before, after *DimensionObservation, relation string) DimensionChange {
	d := DimensionChange{ID: id, Before: before, After: after, Presence: "RETAINED",
		CountDirection: "NOT_COMPARABLE", Regression: "NONE_OBSERVED"}
	if before == nil {
		d.Presence, d.StateTransition, d.ObservationChange = "ADDED", "ABSENT_TO_"+after.Status, "AXIS_ADDED"
		d.Reason = "A new obligation has no prior numeric baseline."
		if after.Status == "FAIL_CLOSED" {
			d.Regression = "FAILURE_OBSERVED"
		}
		return d
	}
	if after == nil {
		d.Presence, d.StateTransition, d.ObservationChange = "REMOVED", before.Status+"_TO_ABSENT", "AXIS_REMOVED"
		d.Regression, d.Reason = "OBLIGATION_REMOVED", "Removing an obligation is explicit; it cannot improve a completeness score."
		return d
	}
	d.StateTransition = before.Status + "_TO_" + after.Status
	d.ObservationChange = observationChange(before, after)
	if before.Status != "FAIL_CLOSED" && after.Status == "FAIL_CLOSED" {
		d.Regression = "FAILURE_OBSERVED"
	}
	if observed(before) && after.Status == "UNKNOWN" {
		d.Regression = "OBSERVATION_LOST"
	}
	if before.Status == "FAIL_CLOSED" && after.Status == "UNKNOWN" {
		d.Regression = "FAILURE_EVIDENCE_LOST"
	}
	switch {
	case relation != "SAME_MEASUREMENT_SCOPE":
		d.Reason = "Profiles or measurement scope are not identical; both observations are retained without a numeric delta."
	case !observed(before) || !observed(after):
		d.Reason = "UNKNOWN and FAIL_CLOSED are not numeric observations."
	case before.Unit != after.Unit || before.Denominator != after.Denominator || before.Denominator == 0:
		d.Reason = "Unit or denominator changed, or no observed denominator exists."
	default:
		d.CountDirection, d.Reason = "UNCHANGED", "Same measurement scope, observed states, unit and nonzero denominator."
		magnitude := 0
		if after.Numerator > before.Numerator {
			d.CountDirection, magnitude = "INCREASED", after.Numerator-before.Numerator
		} else if after.Numerator < before.Numerator {
			d.CountDirection, magnitude, d.Regression = "DECREASED", before.Numerator-after.Numerator, "COUNT_REGRESSION"
		}
		d.CountMagnitude = &magnitude
	}
	return d
}

func observationChange(before, after *DimensionObservation) string {
	switch {
	case before.Status == after.Status:
		return "UNCHANGED_STATE"
	case after.Status == "FAIL_CLOSED":
		return "FAILURE_OBSERVED"
	case before.Status == "FAIL_CLOSED" && after.Status == "UNKNOWN":
		return "FAILURE_NOW_UNKNOWN"
	case before.Status == "FAIL_CLOSED":
		return "FAILURE_REPLACED_BY_OBSERVATION"
	case before.Status == "UNKNOWN" && observed(after):
		return "BECAME_OBSERVED"
	case observed(before) && after.Status == "UNKNOWN":
		return "OBSERVATION_LOST"
	default:
		return "OBSERVED_STATE_CHANGED"
	}
}
