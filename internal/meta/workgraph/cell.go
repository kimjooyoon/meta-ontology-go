package workgraph

func closedCell(gate GateSpec, reason, digest string) Cell {
	return Cell{
		ID: gate.ID, State: "CLOSED", Resolution: "EXACT", Activity: gate.Activity,
		Stage: gate.Stage, Step: gate.Step, Reason: reason, EvidenceKey: gate.EvidenceKey,
		EvidenceDigest: digest, ProofChoice: gate.ProofChoice,
	}
}

func unknownCell(gate GateSpec, resolution, reason string) Cell {
	return Cell{
		ID: gate.ID, State: "UNKNOWN", Resolution: resolution, Activity: gate.Activity,
		Stage: gate.Stage, Step: gate.Step, Reason: reason, EvidenceKey: gate.EvidenceKey,
		ProofChoice: gate.ProofChoice,
	}
}

func refutedCell(gate GateSpec, reason, digest string) Cell {
	return Cell{
		ID: gate.ID, State: "REFUTED", Resolution: "EXACT", Activity: gate.Activity,
		Stage: gate.Stage, Step: gate.Step, Reason: reason, EvidenceKey: gate.EvidenceKey,
		EvidenceDigest: digest, ProofChoice: gate.ProofChoice,
	}
}
