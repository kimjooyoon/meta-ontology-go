package workgraph

func summarize(cells []Cell) Summary {
	summary := Summary{TotalGates: len(cells), RepositoryWrites: 0}
	for _, cell := range cells {
		switch cell.State {
		case "CLOSED":
			summary.ClosedGates++
		case "UNKNOWN":
			summary.UnknownGates++
		case "REFUTED":
			summary.RefutedGates++
		}
	}
	if summary.ClosedGates == summary.TotalGates {
		summary.DischargedClaims = 1
	} else {
		summary.ActiveClaims = 1
	}
	return summary
}

func decision(summary Summary, cells []Cell) (string, string, string, string) {
	if summary.RefutedGates > 0 {
		return "FAIL_CLOSED", "EXACT", "WORKGRAPH_EVIDENCE_REFUTED", "RESOLVE_EVIDENCE_CONTRADICTION"
	}
	if summary.UnknownGates > 0 {
		resolution := "OPERATION_CLASS"
		if cells[0].State == "UNKNOWN" { resolution = "INVARIANT_ONLY" }
		return "FAIL_CLOSED", resolution, "WORKGRAPH_EVIDENCE_UNKNOWN", nextUnknownOperation(cells)
	}
	return "VERTICAL_SLICE_CLOSED", "EXACT", "WORKGRAPH_USER_LOOP_PROVEN", "NONE"
}

func nextUnknownOperation(cells []Cell) string {
	for _, cell := range cells {
		if cell.State != "UNKNOWN" { continue }
		switch cell.ID {
		case "SOURCE_AUTHORITY": return "DECLARE_GOOO_AUTHORITY"
		case "SYNTAX_ACCEPTED": return "RUN_GOOO_CHECK"
		case "META_BOUND": return "BIND_GOOO_META_ACTIVITY"
		case "DETERMINISTIC_REPLAY", "ARTIFACT_GENERATED": return "RUN_GOOO_GENERATE_REPLAY"
		case "RESOURCE_OBSERVED": return "OBSERVE_RESOURCE_SAMPLE"
		default: return "PROVIDE_PREDECESSOR_RECEIPT"
		}
	}
	return "NONE"
}
