package integrationprogress

func evaluatePullObservation(value PullObservation, spec StageSpec) Cell {
	if value.ObservationStatus != "OBSERVED" {
		reason := value.FailureReason
		if reason == "" {
			reason = "PULL_OBSERVATION_UNAVAILABLE"
		}
		return cell(value.Number, spec, StateUnknown, reason, "")
	}
	if value.State != "open" && value.State != "closed" {
		return cell(value.Number, spec, StateRefuted, "PULL_STATE_UNKNOWN", "")
	}
	if !validSHA(value.HeadSHA) {
		return cell(value.Number, spec, StateRefuted, "PULL_HEAD_IDENTITY_INVALID", "")
	}
	if _, ok := parseTime(value.CreatedAt); !ok {
		return cell(value.Number, spec, StateRefuted, "PULL_CREATED_AT_INVALID", "")
	}
	return cell(value.Number, spec, StateClosed, "PULL_OBSERVED", pullRef(value.Number))
}

func pullRef(number int) string {
	return "pull/" + itoa(number)
}
