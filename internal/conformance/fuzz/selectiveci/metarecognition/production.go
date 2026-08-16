package metarecognition

import (
	"math"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/fullsoundness"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func evaluateGooo(c Case) Outcome {
	switch c.Baseline.Subject {
	case SubjectBinding:
		return evaluateBinding(c.Baseline)
	case SubjectGraph:
		return evaluateGraph(c.Baseline)
	case SubjectSoundness:
		return evaluateSoundness(c.Baseline)
	case SubjectPath:
		return evaluatePath(c.Baseline)
	case SubjectResource:
		return evaluateResource(c.Baseline)
	default:
		return productionOutcome(UnknownFullSuiteRequired, ReasonSourceMapRegistry, nil, Work{})
	}
}

func evaluateBinding(b BaselineConfig) Outcome {
	declaration := b.DeclarationName
	if declaration == "" {
		declaration = "Order"
	}
	directive := "//gooo:bind id=\"billing://order\" role=\"HANDWRITTEN_IMPL\"\n"
	if !b.DirectivePresent {
		directive = ""
	}
	source := []byte("package billing\n\n" + directive + "type " + declaration + " struct{}\n")
	sources := []semanticbinding.SourceFile{{Filename: b.ObservedFile, PackagePath: "billing", Source: source}}
	if b.Ambiguous {
		sources = append(sources, sources[0])
	}
	registered := []string{b.StableID}
	if !b.RegistryPresent {
		registered = []string{"billing://other"}
	}
	result, _ := semanticbinding.Extract(semanticbinding.Input{Sources: sources, RegisteredIDs: registered})
	work := Work{Full: len(sources), Selected: len(result.Bindings), ProvRecords: len(result.Bindings) + len(result.Obligations), ProvPaths: len(result.Bindings)}
	if result.Status == semanticbinding.StatusUnknown {
		return productionOutcome(UnknownFullSuiteRequired, ReasonSourceMapRegistry, []string{b.StableID}, work)
	}
	if len(result.Bindings) == 0 {
		return productionOutcome(UnknownFullSuiteRequired, ReasonSourceMapRegistry, []string{b.ObservedFile}, work)
	}
	if len(result.Bindings) != 1 || result.Bindings[0].ID != b.BoundID {
		return productionOutcome(FailClosedUnsound, ReasonBlobWithoutID, []string{b.StableID}, work)
	}
	if b.ExpectedFile != b.ObservedFile || b.ExpectedBlob != b.ObservedBlob {
		return productionOutcome(FailClosedUnsound, ReasonBlobWithoutID, []string{b.StableID}, work)
	}
	reason := ReasonExactBinding
	if result.Bindings[0].DeclarationKey != "Order" {
		reason = ReasonRenameBinding
	}
	return productionOutcome(ClosedSound, reason, []string{b.StableID}, work)
}

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

func soundnessInput(b BaselineConfig) fullsoundness.Input {
	obligationID := b.Obligation.ID
	if obligationID == "" {
		obligationID = "obl-impact"
	}
	authority := fullsoundness.AuthorityAuthoritative
	if b.Obligation.Authority == Candidate {
		authority = fullsoundness.AuthorityCandidate
	}
	obligations := []fullsoundness.Obligation{{ID: obligationID, Authority: authority}}
	commands := make([]fullsoundness.Command, 0, len(b.Commands))
	full, selected := make([]fullsoundness.Outcome, 0, len(b.Commands)), make([]fullsoundness.Outcome, 0, len(b.Commands))
	fullReceipts, selectedReceipts := make([]fullsoundness.ResourceReceipt, 0, len(b.Commands)), make([]fullsoundness.ResourceReceipt, 0, len(b.Commands))
	selectedIDs := make([]string, 0, len(b.Commands))
	for _, command := range b.Commands {
		commandObligation := obligationID
		if command.ID == "cmd-candidate" {
			commandObligation = "obl-candidate"
			if obligationID != commandObligation {
				obligations = append(obligations, fullsoundness.Obligation{ID: commandObligation, Authority: fullsoundness.AuthorityCandidate})
			}
		}
		commands = append(commands, fullsoundness.Command{ID: command.ID, ObligationIDs: []string{commandObligation}, GlobalGuard: command.GlobalGuard})
		full = append(full, fullOutcome(command.ID, command.FullStatus, command.FullDigest))
		fullReceipts = append(fullReceipts, receipt(command.ID, digest("s")))
		if command.Selected {
			selected = append(selected, fullOutcome(command.ID, command.SelectedStatus, command.SelectedDigest))
			selectedReceipts = append(selectedReceipts, receipt(command.ID, digest("s")))
			selectedIDs = append(selectedIDs, command.ID)
		}
	}
	input := fullsoundness.Input{SchemaVersion: fullsoundness.SchemaVersion, SnapshotDigest: digest("s"), PolicyDigest: digest("p"), RegistryDigest: digest("r"), SelectionDigest: digest("l"), ToolchainDigest: digest("t"), RunnerDigest: digest("u"), Obligations: obligations, Commands: commands, ImpactedObligationIDs: []string{obligationID}, SelectedCommandIDs: selectedIDs, SelectionReceipt: &fullsoundness.SelectionReceipt{SnapshotDigest: digest("s"), PolicyDigest: digest("p"), RegistryDigest: digest("r"), SelectionDigest: digest("l"), CommandIDs: selectedIDs}, FullOutcomes: full, SelectedOutcomes: selected, FullResourceReceipts: fullReceipts, SelectedResourceReceipts: selectedReceipts}
	if externalMissing(b.External) {
		if !b.External.Authenticity {
			input.ToolchainDigest = ""
		}
		if !b.External.Provider {
			input.RunnerDigest = ""
		}
		if !b.External.Phase {
			input.SelectionReceipt = nil
		}
		if !b.External.Observer {
			input.FullResourceReceipts, input.SelectedResourceReceipts = nil, nil
		}
	}
	return input
}

func fullOutcome(id string, status Status, seed string) fullsoundness.Outcome {
	result := fullsoundness.Outcome{CommandID: id, OutputDigest: digest(seed)}
	if status == Fail {
		result.Status, result.FailureCode = fullsoundness.OutcomeFail, "FAIL"
	} else {
		result.Status = fullsoundness.OutcomePass
	}
	return result
}

func receipt(id, snapshot string) fullsoundness.ResourceReceipt {
	return fullsoundness.ResourceReceipt{CommandID: id, SnapshotDigest: snapshot, ToolchainDigest: digest("t"), RunnerDigest: digest("u"), CPUCoreNS: 1, AllocatedCPUCount: 1, WallNS: 1, PeakRSSBytes: 1, ReadBytes: 1, WriteBytes: 1}
}

func evaluatePath(b BaselineConfig) Outcome {
	pathID := "path://unknown"
	if len(b.Path.IDs) > 0 {
		pathID = b.Path.IDs[0]
	}
	id := semantic.MustIdentity(pathID)
	requirement := pathclosure.Requirement{PathID: id, RecordIDs: []semantic.ID{semantic.MustIdentity("record://one")}, ExpectedKinds: []semantic.InferenceKind{semantic.InferenceAuthoritativeDeclaration}, StartID: semantic.MustIdentity("node://start"), EndID: semantic.MustIdentity("node://end")}
	requirements := []pathclosure.Requirement{requirement}
	if b.Path.Duplicate {
		requirements = append(requirements, requirement)
	}
	if b.Path.Conflict {
		requirement.RecordIDs = []semantic.ID{semantic.MustIdentity("record://two")}
		requirements = append(requirements, requirement)
	}
	path := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion}
	result := pathclosure.Evaluate(path, requirements)
	work := Work{Units: result.Numerator, Selected: result.Numerator, Full: result.Denominator, ProvRecords: len(path.Edges) + len(path.Evidence), ProvPaths: result.Denominator}
	if result.Status == pathclosure.FAIL_CLOSED && b.Path.Conflict {
		return productionOutcome(FailClosedUnsound, ReasonConflictingReceipt, []string{id.String()}, work)
	}
	if result.Status == pathclosure.FAIL_CLOSED {
		return productionOutcome(FailClosedUnsound, ReasonDuplicateReceipt, []string{id.String()}, work)
	}
	return productionOutcome(UnknownFullSuiteRequired, ReasonExternalMissing, []string{id.String()}, work)
}

func evaluateResource(b BaselineConfig) Outcome {
	samples := make([]resourceenvelope.Sample, 6)
	for index := range samples {
		samples[index] = resourceenvelope.Sample{CPUCoreNS: math.MaxUint64, WallNS: 1}
	}
	envelope := resourceenvelope.Envelope{SchemaVersion: resourceenvelope.SchemaVersion, RunnerImageDigest: digest("r"), AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 5, Limits: resourceenvelope.Limits{CPUCoreNS: math.MaxUint64, PeakRSSBytes: math.MaxUint64, ReadBytes: math.MaxUint64, WriteBytes: math.MaxUint64}, Samples: samples}
	result := resourceenvelope.Evaluate(envelope)
	work := Work{Units: int(envelope.SampleCount), Selected: int(envelope.SampleCount), Full: len(envelope.Samples), ProvRecords: len(envelope.Samples), ProvPaths: 1}
	if result.Status != resourceenvelope.PASS {
		return productionOutcome(UnknownFullSuiteRequired, ReasonInvalidResource, []string{"receipt-1"}, work)
	}
	return productionOutcome(ClosedSound, ReasonExactBinding, nil, work)
}

func commandIDs(values []CommandAssertion, include func(CommandAssertion) bool) []string {
	ids := make([]string, 0, len(values))
	for _, value := range values {
		if include(value) {
			ids = append(ids, value.ID)
		}
	}
	return sorted(ids)
}

func externalMissing(value ExternalAssertion) bool {
	return !(value.Authenticity && value.Provider && value.Phase && value.Observer)
}

func externalInputID(value ExternalAssertion) string {
	if !value.Authenticity {
		return "external-authenticity"
	}
	if !value.Provider {
		return "external-provider"
	}
	if !value.Phase {
		return "external-phase"
	}
	return "external-observer"
}

func productionOutcome(state State, reason Reason, ids []string, work Work) Outcome {
	work.Units = work.Selected
	return Outcome{State: state, Reason: reason, LocalizedIDs: sorted(ids), Work: work}
}
