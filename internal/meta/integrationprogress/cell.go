package integrationprogress

const (
	StateClosed  = "CLOSED"
	StateOpen    = "OPEN"
	StateUnknown = "UNKNOWN"
	StateRefuted = "REFUTED"
)

func cell(number int, spec StageSpec, state, reason, evidence string) Cell {
	resolution := "EXACT"
	if state == StateUnknown {
		resolution = "LOWER_RESOLUTION"
	} else if state == StateRefuted {
		resolution = "INVARIANT_ONLY"
	}
	return Cell{PullRequest: number, Stage: spec.ID, Step: spec.Step,
		State: state, Reason: reason, Resolution: resolution, EvidenceRef: evidence}
}

func dependentCell(number int, spec StageSpec, upstream Cell) Cell {
	switch upstream.State {
	case StateOpen:
		return cell(number, spec, StateOpen, "UPSTREAM_OPEN", "")
	case StateUnknown:
		return cell(number, spec, StateUnknown, "UPSTREAM_UNKNOWN", "")
	default:
		return cell(number, spec, StateRefuted, "UPSTREAM_REFUTED", "")
	}
}
