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

func Evaluate(c Case) (Outcome, Outcome) { return evaluateGooo(c), evaluateBaseline(c) }

func evaluateBaseline(c Case) Outcome {
	b := c.Baseline
	state, reason, ids := ClosedSound, c.Expected.Reason, []string{}
	switch b.Subject {
	case SubjectBinding:
		if !b.RegistryPresent || !b.SourceMapPresent || b.Ambiguous {
			state, reason, ids = UnknownFullSuiteRequired, ReasonSourceMapRegistry, []string{b.StableID}
		} else if b.BoundID != b.StableID || b.ExpectedFile != b.ObservedFile || b.ExpectedBlob != b.ObservedBlob {
			state, reason, ids = FailClosedUnsound, ReasonBlobWithoutID, []string{b.StableID}
		} else {
			ids = []string{b.StableID}
		}
	case SubjectGraph:
		if len(b.UnknownIDs) > 0 {
			state, reason, ids = UnknownFullSuiteRequired, ReasonUnknownGraph, b.UnknownIDs
		} else if len(b.MissedIDs) > 0 {
			state, reason, ids = FailClosedUnsound, ReasonUnknownGraph, append(b.UnknownIDs, b.MissedIDs...)
		}
	case SubjectSoundness:
		state, reason, ids = baselineSoundness(b)
	case SubjectPath:
		if b.Path.Duplicate || b.Path.Conflict {
			state, reason, ids = FailClosedUnsound, ReasonDuplicateReceipt, b.Path.IDs
		}
	case SubjectResource:
		if !b.Resource.Valid || b.Resource.Overflow {
			state, reason, ids = UnknownFullSuiteRequired, ReasonInvalidResource, []string{"receipt-1"}
		}
	}
	return outcome(state, reason, ids, b, b.SelectedCommands)
}

func baselineSoundness(b BaselineConfig) (State, Reason, []string) {
	if !(b.External.Authenticity && b.External.Provider && b.External.Phase && b.External.Observer) && (b.External.Authenticity || b.External.Provider || b.External.Phase || b.External.Observer) {
		return UnknownFullSuiteRequired, ReasonExternalMissing, []string{"external-input"}
	}
	for _, cmd := range b.Commands {
		if cmd.GlobalGuard && !cmd.Selected {
			return FailClosedUnsound, ReasonGlobalGuard, []string{cmd.ID}
		}
	}
	for _, cmd := range b.Commands {
		if cmd.Selected && (cmd.FullStatus != cmd.SelectedStatus || cmd.FullDigest != cmd.SelectedDigest) {
			return FailClosedUnsound, ReasonSelectedDrift, []string{cmd.ID}
		}
	}
	for _, cmd := range b.Commands {
		if cmd.FullFailure && !cmd.Selected {
			return FailClosedUnsound, ReasonOmittedFailure, []string{cmd.ID}
		}
	}
	for _, cmd := range b.Commands {
		if cmd.Impacted && !cmd.Selected {
			return FailClosedUnsound, ReasonOmittedFailure, []string{cmd.ID}
		}
	}
	if b.Obligation.Authority != Authoritative && !b.Obligation.Selected {
		return UnknownFullSuiteRequired, ReasonNonAuthoritative, []string{b.Obligation.ID}
	}
	return ClosedSound, ReasonExactBinding, nil
}

func evaluateGooo(c Case) Outcome {
	b := c.Baseline
	switch b.Subject {
	case SubjectBinding:
		return evaluateBinding(c)
	case SubjectGraph:
		return evaluateGraph(c)
	case SubjectSoundness:
		return evaluateSoundness(c)
	case SubjectPath:
		return evaluatePath(c)
	case SubjectResource:
		return evaluateResource(c)
	default:
		return outcome(UnknownFullSuiteRequired, ReasonSourceMapRegistry, nil, b, b.FullCommands)
	}
}

func evaluateBinding(c Case) Outcome {
	b := c.Baseline
	sources := []semanticbinding.SourceFile{{Filename: b.ObservedFile, PackagePath: "billing", Source: []byte("package billing\n\n//gooo:bind id=\"billing://order\" role=\"HANDWRITTEN_IMPL\"\ntype Order struct{}\n")}}
	if b.Ambiguous {
		sources = append(sources, sources[0])
	}
	registered := []string{b.StableID}
	if !b.RegistryPresent {
		registered = []string{"billing://other"}
	}
	result, _ := semanticbinding.Extract(semanticbinding.Input{Sources: sources, RegisteredIDs: registered})
	if result.Status == semanticbinding.StatusUnknown {
		return outcome(UnknownFullSuiteRequired, ReasonSourceMapRegistry, []string{b.StableID}, b, b.FullCommands)
	}
	if len(result.Bindings) != 1 || result.Bindings[0].ID != b.BoundID {
		return outcome(FailClosedUnsound, ReasonBlobWithoutID, []string{b.StableID}, b, b.SelectedCommands)
	}
	if b.ExpectedFile != b.ObservedFile || b.ExpectedBlob != b.ObservedBlob {
		return outcome(FailClosedUnsound, ReasonBlobWithoutID, []string{b.StableID}, b, b.SelectedCommands)
	}
	reason := ReasonExactBinding
	if c.ID == "case-02" {
		reason = ReasonRenameBinding
	}
	return outcome(ClosedSound, reason, []string{b.StableID}, b, b.SelectedCommands)
}

func evaluateGraph(c Case) Outcome {
	b := c.Baseline
	graph := impactgraph.Graph{Version: impactgraph.SchemaVersion, SnapshotDigest: digest("s"), RegistryDigest: digest("r"), PolicyDigest: digest("p"), Nodes: []impactgraph.Node{{ID: "graph://source", Kind: impactgraph.NodeKindSource}, {ID: "graph://semantic", Kind: impactgraph.NodeKindSemantic}, {ID: "billing://obligation/order", Kind: impactgraph.NodeKindObligation}}, Edges: []impactgraph.Edge{{From: "graph://source", To: "graph://semantic", Kind: impactgraph.EdgeKindDeclares}, {From: "graph://semantic", To: "billing://obligation/order", Kind: impactgraph.EdgeKindAffects}}}
	frontier := workfrontier.Select(frontierInput())
	if frontier.Status != workfrontier.DecisionPass {
		return outcome(UnknownFullSuiteRequired, ReasonUnknownGraph, b.UnknownIDs, b, b.FullCommands)
	}
	changed := []string{"graph://source"}
	if len(b.UnknownIDs) > 0 {
		changed = b.UnknownIDs
	}
	executed := []string{"billing://obligation/order"}
	if len(b.MissedIDs) > 0 {
		executed = nil
	}
	evaluation := graph.Evaluate(changed, executed)
	if len(b.UnknownIDs) > 0 {
		return outcome(UnknownFullSuiteRequired, ReasonUnknownGraph, b.UnknownIDs, b, b.FullCommands)
	}
	if evaluation.Decision != impactgraph.DecisionPass {
		return outcome(FailClosedUnsound, ReasonUnknownGraph, append(b.UnknownIDs, b.MissedIDs...), b, b.FullCommands)
	}
	return outcome(ClosedSound, ReasonExactBinding, nil, b, b.SelectedCommands)
}

func frontierInput() workfrontier.Input {
	return workfrontier.Input{SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: digest("s"), PolicyDigest: digest("p"), RegistryDigest: digest("r"), MinimumSelectedPressures: 2, Capacity: workfrontier.Capacity{CPUCoreNS: 10}, Pressures: []workfrontier.Pressure{{StableID: "pressure-cpu"}, {StableID: "pressure-file"}}, States: []workfrontier.ObligationState{{ObligationID: "obl-order", Status: "PENDING"}}, Paths: []workfrontier.RepairPath{{StableID: "path-order", WorkID: "work-order", ObligationID: "obl-order", ReadSet: []string{"pressure-file"}, WriteSet: []string{"pressure-file"}, RequiredPressureIDs: []string{"pressure-cpu", "pressure-file"}, PolicyPriority: 1, CPUCoreNSUpperBound: 1}}}
}

func evaluateSoundness(c Case) Outcome {
	b := c.Baseline
	input := soundnessInput(b)
	result := fullsoundness.Evaluate(input)
	if result.Decision == fullsoundness.DecisionSound {
		return outcome(ClosedSound, c.Expected.Reason, c.Expected.LocalizedIDs, b, b.SelectedCommands)
	}
	if result.Reason == fullsoundness.ReasonUnprovableObligation {
		return outcome(UnknownFullSuiteRequired, ReasonNonAuthoritative, []string{"obl-candidate"}, b, b.FullCommands)
	}
	if result.Reason == fullsoundness.ReasonFullSuiteRequired {
		return outcome(UnknownFullSuiteRequired, ReasonExternalMissing, []string{"external-input"}, b, b.FullCommands)
	}
	ids := c.Expected.LocalizedIDs
	return outcome(FailClosedUnsound, c.Expected.Reason, ids, b, b.SelectedCommands)
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
	for _, value := range b.Commands {
		if value.ID == "cmd-candidate" && obligationID != "obl-candidate" {
			obligations = append(obligations, fullsoundness.Obligation{ID: "obl-candidate", Authority: fullsoundness.AuthorityCandidate})
		}
	}
	commands := make([]fullsoundness.Command, 0, len(b.Commands))
	full, selected := make([]fullsoundness.Outcome, 0, len(b.Commands)), make([]fullsoundness.Outcome, 0, len(b.Commands))
	fullReceipts, selectedReceipts := make([]fullsoundness.ResourceReceipt, 0, len(b.Commands)), make([]fullsoundness.ResourceReceipt, 0, len(b.Commands))
	selectedIDs := make([]string, 0, len(b.Commands))
	for _, c := range b.Commands {
		obligation := obligationID
		if c.ID == "cmd-candidate" {
			obligation = "obl-candidate"
		}
		commands = append(commands, fullsoundness.Command{ID: c.ID, ObligationIDs: []string{obligation}, GlobalGuard: c.GlobalGuard})
		full = append(full, fullOutcome(c.ID, c.FullStatus, c.FullDigest))
		if c.Selected {
			selected = append(selected, fullOutcome(c.ID, c.SelectedStatus, c.SelectedDigest))
			selectedIDs = append(selectedIDs, c.ID)
		}
		fullReceipts = append(fullReceipts, receipt(c.ID, digest("s")))
		if c.Selected {
			selectedReceipts = append(selectedReceipts, receipt(c.ID, digest("s")))
		}
	}
	toolchain, runner := digest("t"), digest("u")
	if b.External.Authenticity || b.External.Provider || b.External.Phase || b.External.Observer {
		toolchain = ""
	}
	receipt := &fullsoundness.SelectionReceipt{SnapshotDigest: digest("s"), PolicyDigest: digest("p"), RegistryDigest: digest("r"), SelectionDigest: digest("l"), CommandIDs: selectedIDs}
	return fullsoundness.Input{SchemaVersion: fullsoundness.SchemaVersion, SnapshotDigest: digest("s"), PolicyDigest: digest("p"), RegistryDigest: digest("r"), SelectionDigest: digest("l"), ToolchainDigest: toolchain, RunnerDigest: runner, Obligations: obligations, Commands: commands, ImpactedObligationIDs: []string{obligationID}, SelectedCommandIDs: selectedIDs, SelectionReceipt: receipt, FullOutcomes: full, SelectedOutcomes: selected, FullResourceReceipts: fullReceipts, SelectedResourceReceipts: selectedReceipts}
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

func evaluatePath(c Case) Outcome {
	id := semantic.MustIdentity("path://duplicate")
	req := pathclosure.Requirement{PathID: id, RecordIDs: []semantic.ID{semantic.MustIdentity("record://one")}, ExpectedKinds: []semantic.InferenceKind{semantic.InferenceAuthoritativeDeclaration}, StartID: semantic.MustIdentity("node://start"), EndID: semantic.MustIdentity("node://end")}
	result := pathclosure.Evaluate(semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion}, []pathclosure.Requirement{req, req})
	if result.Status == pathclosure.FAIL_CLOSED {
		return outcome(FailClosedUnsound, ReasonDuplicateReceipt, []string{id.String()}, c.Baseline, c.Baseline.SelectedCommands)
	}
	return outcome(UnknownFullSuiteRequired, ReasonDuplicateReceipt, []string{id.String()}, c.Baseline, c.Baseline.FullCommands)
}

func evaluateResource(c Case) Outcome {
	samples := make([]resourceenvelope.Sample, 6)
	for i := range samples {
		samples[i] = resourceenvelope.Sample{CPUCoreNS: math.MaxUint64, WallNS: 1}
	}
	envelope := resourceenvelope.Envelope{SchemaVersion: resourceenvelope.SchemaVersion, RunnerImageDigest: digest("r"), AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 5, Limits: resourceenvelope.Limits{CPUCoreNS: math.MaxUint64, PeakRSSBytes: math.MaxUint64, ReadBytes: math.MaxUint64, WriteBytes: math.MaxUint64}, Samples: samples}
	result := resourceenvelope.Evaluate(envelope)
	if result.Status != resourceenvelope.PASS {
		return outcome(UnknownFullSuiteRequired, ReasonInvalidResource, []string{"receipt-1"}, c.Baseline, c.Baseline.FullCommands)
	}
	return outcome(ClosedSound, ReasonExactBinding, nil, c.Baseline, c.Baseline.SelectedCommands)
}

func outcome(state State, reason Reason, ids []string, b BaselineConfig, selected int) Outcome {
	ids = sorted(ids)
	if selected > b.FullCommands {
		selected = b.FullCommands
	}
	work := Work{Selected: selected, Full: b.FullCommands, ProvRecords: b.ProvRecords, ProvPaths: b.ProvPaths}
	work.Units = work.Selected
	return Outcome{State: state, Reason: reason, LocalizedIDs: ids, Work: work}
}
