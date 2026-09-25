package reproducibilitysemantics

type transitionWitness struct {
	Channel    string          `json:"channel"`
	From       string          `json:"from"`
	To         string          `json:"to"`
	Coordinate Coordinate      `json:"coordinate"`
	Byte       Evidence        `json:"byte"`
	Meaning    MeaningEvidence `json:"meaning"`
}

func claimTransitions(item Case) (Transition, Transition, Transition) {
	byte := makeTransition("byte", item.Byte.Status, item.Byte, MeaningEvidence{})
	meaning := makeTransition("meaning", item.Meaning.Status, Evidence{}, item.Meaning)
	joint := makeTransition("joint", item.Status, item.Byte, item.Meaning)
	return byte, meaning, joint
}

func makeTransition(channel, to string, byte Evidence, meaning MeaningEvidence) Transition {
	coordinate := coordinate(statusInt(to == StatusDischarged), 1)
	reason := "OPEN_TO_" + to
	if to == StatusOpen {
		reason = "EVIDENCE_REMAINS_OPEN"
	}
	witness := transitionWitness{Channel: channel, From: StatusOpen, To: to, Coordinate: coordinate,
		Byte: byte, Meaning: meaning}
	return Transition{From: StatusOpen, To: to, Coordinate: coordinate, Stage: "transition", Step: channel,
		Reason: reason, EvidenceDigest: digestValue(witness)}
}
