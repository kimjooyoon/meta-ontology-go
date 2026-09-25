package workgraph

func buildCells(contract Contract, observation Observation) []Cell {
	cells := make([]Cell, 0, GateCount)
	for _, gate := range contract.Gates {
		if gate.ID == "USER_ROUNDTRIP" {
			cells = append(cells, roundtripCell(contract, observation, gate, cells))
			continue
		}
		cells = append(cells, evidenceCell(contract, observation, gate))
	}
	return cells
}

func evidenceCell(contract Contract, observation Observation, gate GateSpec) Cell {
	switch gate.ID {
	case "SOURCE_AUTHORITY":
		if observation.SourceDigest == "" || observation.SourcePath != contract.Source {
			return unknownCell(gate, "INVARIANT_ONLY", "SOURCE_AUTHORITY_NOT_OBSERVED")
		}
		return closedCell(gate, "GOOO_SOURCE_AUTHORITY_OBSERVED", observation.SourceDigest)
	case "SYNTAX_ACCEPTED":
		if observation.CheckDigest == "" {
			return unknownCell(gate, "OPERATION_CLASS", "GOOO_CHECK_RECEIPT_MISSING")
		}
		return closedCell(gate, "GOOO_SYNTAX_ACCEPTED", observation.CheckDigest)
	case "META_BOUND":
		bound, missing := sourceBinding(contract, observation.SourceText)
		if observation.SourceText == "" {
			return unknownCell(gate, "INVARIANT_ONLY", "GOOO_SOURCE_NOT_READ")
		}
		if !bound {
			return refutedCell(gate, "GOOO_META_DECLARATION_MISSING:"+missing, observation.SourceDigest)
		}
		return closedCell(gate, "GOOO_META_ACTIVITIES_BOUND", observation.SourceDigest)
	case "DETERMINISTIC_REPLAY":
		return replayCell(observation, gate)
	case "ARTIFACT_GENERATED":
		return artifactCell(observation, gate)
	default:
		return resourceCell(observation, gate)
	}
}
