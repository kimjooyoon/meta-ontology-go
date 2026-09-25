package workgraph

func indicators(report Report) []Indicator {
	closed := int64(report.Summary.ClosedGates)
	total := int64(report.Summary.TotalGates)
	replay := indicatorCellValue(report.Cells, "DETERMINISTIC_REPLAY")
	resource := indicatorCellValue(report.Cells, "RESOURCE_OBSERVED")
	roundtrip := indicatorCellValue(report.Cells, "USER_ROUNDTRIP")
	trace := int64(0)
	if report.Claim.TraceRetained && report.Claim.Before.State == "UNKNOWN" {
		trace = 1
	}
	return []Indicator{
		indicator("gooo.metric.workgraph.closed-gates.v1", "OUTCOME", closed, total, "cells", ">=", total, "CloseProjectClaim", "COHERENCE"),
		indicator("gooo.metric.workgraph.replay.v1", "DRIVER", replay, 1, "replays", ">=", 1, "ReplayProjectGeneration", "REGRESSION"),
		indicator("gooo.metric.workgraph.resource-observation.v1", "DRIVER", resource, 1, "samples", ">=", 1, "ObserveProjectResources", "REGRESSION"),
		indicator("gooo.metric.workgraph.user-roundtrip.v1", "OUTCOME", roundtrip, 1, "roundtrips", ">=", 1, "CloseProjectClaim", "COHERENCE"),
		indicator("gooo.metric.workgraph.unknown-trace.v1", "GUARDRAIL", trace, 1, "traces", ">=", 1, "LowerProjectResolution", "FOUNDATION"),
		indicator("gooo.metric.workgraph.repository-writes.v1", "GUARDRAIL", 0, 1, "writes", "<=", 0, "ObserveProject", "FOUNDATION"),
	}
}

func indicator(id, class string, value, total int64, unit, relation string, target int64, activity, proof string) Indicator {
	state := "GAP"
	if relation == ">=" && value >= target || relation == "<=" && value <= target {
		state = "SATISFIED"
	}
	return Indicator{ID: id, Class: class, Value: value, Total: total, Unit: unit, Relation: relation, Target: target, Activity: activity, ProofChoice: proof, State: state}
}

func indicatorCellValue(cells []Cell, id string) int64 {
	for _, cell := range cells {
		if cell.ID == id && cell.State == "CLOSED" {
			return 1
		}
	}
	return 0
}
