package workgraph

import "fmt"

func roundtripCell(contract Contract, observation Observation, gate GateSpec, cells []Cell) Cell {
	if hasState(cells, "REFUTED") {
		return refutedCell(gate, "ROUNDTRIP_PREDECESSOR_GATE_REFUTED", observation.PredecessorDigest)
	}
	if !allClosed(cells) {
		return unknownCell(gate, "OPERATION_CLASS", "ROUNDTRIP_PREDECESSOR_GATES_UNKNOWN")
	}
	if observation.Predecessor == nil {
		return unknownCell(gate, "OPERATION_CLASS", "PREDECESSOR_UNKNOWN_RECEIPT_MISSING")
	}
	if err := validatePredecessor(contract, observation); err != nil {
		return refutedCell(gate, "PREDECESSOR_RECEIPT_INVALID:"+err.Error(), observation.PredecessorDigest)
	}
	return closedCell(gate, "UNKNOWN_CLAIM_DISCHARGED_BY_EVIDENCE", observation.PredecessorDigest)
}

func validatePredecessor(contract Contract, observation Observation) error {
	prior := observation.Predecessor
	if prior.Schema != ReportSchema || prior.HeadSHA != observation.HeadSHA {
		return fmt.Errorf("identity mismatch")
	}
	if prior.ContractDigest != DigestValue(contract) || prior.SourceDigest != observation.SourceDigest {
		return fmt.Errorf("authority mismatch")
	}
	if prior.Decision != "FAIL_CLOSED" || prior.Summary.UnknownGates == 0 {
		return fmt.Errorf("unknown state missing")
	}
	if prior.NextOperation != "RUN_GOOO_GENERATE_REPLAY" || prior.MutationAllowed {
		return fmt.Errorf("operation mismatch")
	}
	if prior.Claim.After.Status != "ACTIVE" || prior.Claim.After.State != "UNKNOWN" {
		return fmt.Errorf("claim state mismatch")
	}
	return nil
}

func allClosed(cells []Cell) bool { return !hasState(cells, "UNKNOWN") && !hasState(cells, "REFUTED") }
func hasState(cells []Cell, state string) bool {
	for _, cell := range cells {
		if cell.State == state {
			return true
		}
	}
	return false
}
