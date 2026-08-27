package languagesourcebindingpromotion

type component struct {
	Status       string
	UnknownClass string
	Reason       string
	Coordinate   Coordinate
	Evidence     []string
}

func discharged(reason, stage, step string, evidence ...string) component {
	return component{Status: "DISCHARGED", Reason: reason,
		Coordinate: Coordinate{Stage: stage, Step: step, Reason: reason}, Evidence: evidence}
}

func open(reason, stage, step string, evidence ...string) component {
	return component{Status: "OPEN", UnknownClass: "DIRECT_MISSING", Reason: reason,
		Coordinate: Coordinate{Stage: stage, Step: step, Reason: reason}, Evidence: evidence}
}

func refuted(reason, stage, step string, evidence ...string) component {
	return component{Status: "REFUTED", Reason: reason,
		Coordinate: Coordinate{Stage: stage, Step: step, Reason: reason}, Evidence: evidence}
}

func claim(id string, value component) ClaimResult {
	return ClaimResult{ID: id, Status: value.Status, UnknownClass: value.UnknownClass,
		Reason: value.Reason, Coordinate: value.Coordinate, EvidenceDigests: value.Evidence}
}
