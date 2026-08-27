package workgraph

func claimLifecycle(contract Contract, observation Observation, cells []Cell, summary Summary, next string) ClaimLifecycle {
	after := claimSnapshot(cells, summary, next)
	before := after
	if observation.Predecessor != nil {
		before = observation.Predecessor.Claim.After
	}
	return ClaimLifecycle{
		ID: contract.Claim.ID, Entity: contract.Claim.Entity, Before: before, After: after,
		TraceRetained: true, PredecessorDigest: observation.PredecessorDigest,
	}
}

func claimSnapshot(cells []Cell, summary Summary, next string) ClaimSnapshot {
	if summary.ClosedGates == summary.TotalGates {
		return ClaimSnapshot{
			Status: "DISCHARGED", State: "CLOSED", Resolution: "EXACT",
			Reason: "CLAIM_DISCHARGED_EVIDENCE_TRACE_RETAINED", NextOperation: "NONE",
		}
	}
	for _, cell := range cells {
		if cell.State == "CLOSED" { continue }
		status := "ACTIVE"
		if cell.State == "REFUTED" { status = "CONTESTED" }
		return ClaimSnapshot{
			Status: status, State: cell.State, Resolution: cell.Resolution,
			Stage: cell.Stage, Step: cell.Step, Reason: cell.Reason, NextOperation: next,
		}
	}
	return ClaimSnapshot{Status: "ACTIVE", State: "UNKNOWN", Resolution: "INVARIANT_ONLY", Reason: "CLAIM_STATE_MISSING", NextOperation: next}
}
