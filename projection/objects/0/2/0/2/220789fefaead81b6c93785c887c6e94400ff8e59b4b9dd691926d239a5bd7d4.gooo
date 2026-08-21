package metarecognition

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/fullsoundness"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

func evaluateGraph(b BaselineConfig) Outcome {
	graph := impactgraph.Graph{Version: impactgraph.SchemaVersion, SnapshotDigest: digest("s"), RegistryDigest: digest("r"), PolicyDigest: digest("p"), Nodes: []impactgraph.Node{{ID: "graph://source", Kind: impactgraph.NodeKindSource}, {ID: "graph://semantic", Kind: impactgraph.NodeKindSemantic}, {ID: "billing://obligation/order", Kind: impactgraph.NodeKindObligation}}, Edges: []impactgraph.Edge{{From: "graph://source", To: "graph://semantic", Kind: impactgraph.EdgeKindDeclares}, {From: "graph://semantic", To: "billing://obligation/order", Kind: impactgraph.EdgeKindAffects}}}
	frontier := workfrontier.Select(frontierInput())
	changed := []string{"graph://source"}
	if len(b.UnknownIDs) > 0 {
		changed = b.UnknownIDs
	}
	executed := []string{"billing://obligation/order"}
	if len(b.MissedIDs) > 0 {
		executed = nil
	}
	evaluation := graph.Evaluate(changed, executed)
	work := Work{Full: len(graph.Nodes), Selected: len(frontier.SelectedIDs), ProvRecords: len(graph.Nodes) + len(graph.Edges), ProvPaths: len(evaluation.Required)}
	if frontier.Status != workfrontier.DecisionPass || evaluation.Decision == impactgraph.DecisionUnknown {
		return productionOutcome(UnknownFullSuiteRequired, ReasonUnknownGraph, b.UnknownIDs, work)
	}
	if evaluation.FailureCode == impactgraph.FailureCodeMissedObligations {
		return productionOutcome(FailClosedUnsound, ReasonMissedObligation, evaluation.Missed, work)
	}
	if evaluation.Decision != impactgraph.DecisionPass {
		return productionOutcome(FailClosedUnsound, ReasonUnknownGraph, evaluation.Missed, work)
	}
	return productionOutcome(ClosedSound, ReasonExactBinding, nil, work)
}
func frontierInput() workfrontier.Input {
	return workfrontier.Input{SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: digest("s"), PolicyDigest: digest("p"), RegistryDigest: digest("r"), MinimumSelectedPressures: 2, Capacity: workfrontier.Capacity{CPUCoreNS: 10}, Pressures: []workfrontier.Pressure{{StableID: "pressure-cpu"}, {StableID: "pressure-file"}}, States: []workfrontier.ObligationState{{ObligationID: "obl-order", Status: "PENDING"}}, Paths: []workfrontier.RepairPath{{StableID: "path-order", WorkID: "work-order", ObligationID: "obl-order", ReadSet: []string{"pressure-file"}, WriteSet: []string{"pressure-file"}, RequiredPressureIDs: []string{"pressure-cpu", "pressure-file"}, PolicyPriority: 1, CPUCoreNSUpperBound: 1}}}
}
func evaluateSoundness(b BaselineConfig) Outcome {
	input := soundnessInput(b)
	result := fullsoundness.Evaluate(input)
	work := soundnessWork(input, result)
	if result.Decision == fullsoundness.DecisionSound {
		return productionOutcome(ClosedSound, ReasonExactBinding, nil, work)
	}
	if result.Reason == fullsoundness.ReasonUnprovableObligation {
		return productionOutcome(UnknownFullSuiteRequired, ReasonNonAuthoritative, []string{b.Obligation.ID}, work)
	}
	if result.Reason == fullsoundness.ReasonFullSuiteRequired {
		return productionOutcome(UnknownFullSuiteRequired, ReasonExternalMissing, []string{externalInputID(b.External)}, work)
	}
	if result.Reason == fullsoundness.ReasonGlobalGuardOmitted {
		return productionOutcome(FailClosedUnsound, ReasonGlobalGuard, commandIDs(b.Commands, func(value CommandAssertion) bool { return value.GlobalGuard && !value.Selected }), work)
	}
	if result.Reason == fullsoundness.ReasonOmittedFullFailure {
		return productionOutcome(FailClosedUnsound, ReasonOmittedFailure, commandIDs(b.Commands, func(value CommandAssertion) bool { return value.FullFailure && !value.Selected }), work)
	}
	return productionOutcome(FailClosedUnsound, ReasonSelectedDrift, commandIDs(b.Commands, func(value CommandAssertion) bool {
		return value.Selected && (value.FullStatus != value.SelectedStatus || value.FullDigest != value.SelectedDigest)
	}), work)
}
func soundnessWork(input fullsoundness.Input, result fullsoundness.Output) Work {
	return Work{Units: int(result.SelectedCommandCount), Selected: int(result.SelectedCommandCount), Full: int(result.CommandCount), ProvRecords: len(input.Obligations) + len(input.Commands) + len(input.FullOutcomes) + len(input.FullResourceReceipts), ProvPaths: len(input.SelectedCommandIDs)}
}
