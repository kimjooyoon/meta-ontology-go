package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type parsedClaimDependencyActivity struct {
	declaration *syntax.ActivityDecl
	program     claimDependencyProgram
}

func resolveClaimDependencies(sourceFile string, source []byte, file *syntax.File) claimDependencyReport {
	report := newClaimDependencyReport(sourceFile, source, file)
	activities := make([]*syntax.ActivityDecl, 0)
	producers := make(map[string]*syntax.ActivityDecl)
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		report.Summary.ActivitiesTotal++
		if prior, duplicate := producers[activity.Output]; duplicate {
			return refuteClaimDependencies(report, "META_BINDING", "BIND_OUTPUT_PRODUCERS", "CLAIM_OUTPUT_PRODUCER_AMBIGUOUS", "RESTORE_ONE_PRODUCER_PER_ENTITY", claimDependencyRegression, []string{prior.Name, activity.Name})
		}
		producers[activity.Output] = activity
		activities = append(activities, activity)
	}
	if len(activities) == 0 {
		return unknownClaimDependencies(report, "CLAIM_DEPENDENCY", "OBSERVE_CLAIM_ACTIVITIES", "CLAIM_DEPENDENCY_ACTIVITIES_UNAVAILABLE", "DIRECT_MISSING", "DECLARE_CLAIM_DEPENDENCY_ACTIVITIES", claimDependencyFoundation, []string{})
	}
	parsed := make([]parsedClaimDependencyActivity, 0, len(activities))
	kinds := make(map[string]bool, len(claimDependencyEdgeKinds))
	for index, activity := range activities {
		if !activity.ValueProgramPresent {
			return refuteClaimDependencies(report, "CLAIM_DEPENDENCY", "OBSERVE_VALUE_PROGRAM", "CLAIM_DEPENDENCY_PROGRAM_MISSING", "DECLARE_CLAIM_DEPENDENCY_VALUE_PROGRAM", claimDependencyFoundation, []string{activity.Name})
		}
		program, failed := parseClaimDependencyProgram(activity.ValueProgram)
		if failed != nil {
			return refuteClaimDependencies(report, "CLAIM_DEPENDENCY", "PARSE_VALUE_PROGRAM", failed.reason, failed.next, claimDependencyRegression, []string{activity.Name})
		}
		proof := claimDependencyCoherence
		if program.role == claimDependencyRootRole {
			proof = claimDependencyFoundation
			report.Summary.RecoverableRoots++
		} else {
			report.Summary.TypedDeclarations++
			kinds[program.kind] = true
		}
		report.Summary.ActivitiesObserved++
		report.Nodes = append(report.Nodes, claimDependencyNode{
			Ordinal: index + 1, Activity: activity.Name, OutputEntity: activity.Output, Role: program.role,
			Label: program.label, ProofChoice: proof, ValueProgramDigest: claimResolutionDigest([]byte(activity.ValueProgram)),
		})
		parsed = append(parsed, parsedClaimDependencyActivity{declaration: activity, program: program})
	}
	report.Summary.EdgeKindsObserved = len(kinds)
	for _, item := range parsed {
		activity := item.declaration
		if item.program.role == claimDependencyRootRole {
			for _, input := range activity.Inputs {
				if producer, found := producers[input.Name]; found {
					return refuteClaimDependencies(report, "FOUNDATION", "VERIFY_RECOVERABLE_ROOT", "RECOVERABLE_ROOT_HAS_CLAIM_PREDECESSOR", "REMOVE_PREDECESSOR_OR_SELECT_COHERENCE", claimDependencyRegression, []string{producer.Name, activity.Name})
				}
			}
			continue
		}
		if len(activity.Inputs) == 0 {
			return refuteClaimDependencies(report, "CLAIM_DEPENDENCY", "BIND_INPUT_PRODUCERS", "TYPED_DEPENDENCY_INPUT_MISSING", "DECLARE_DEPENDENCY_INPUT", claimDependencyRegression, []string{activity.Name})
		}
		for _, input := range activity.Inputs {
			report.Summary.DependencyInputs++
			producer, found := producers[input.Name]
			if !found {
				report.Gaps = append(report.Gaps, claimDependencyGap{
					Activity: activity.Name, InputEntity: input.Name, Stage: "DEPENDENCY_DISCOVERY", Step: "BIND_INPUT_PRODUCER",
					Reason: "CLAIM_INPUT_PRODUCER_UNAVAILABLE", UnknownClass: "DIRECT_MISSING", NextOperation: "DECLARE_INPUT_PRODUCER",
				})
				continue
			}
			edgeOrdinal := len(report.Edges) + 1
			report.Edges = append(report.Edges, claimDependencyEdge{
				Ordinal: edgeOrdinal, ID: fmt.Sprintf("gooo://claim-dependency/edge/%s/%03d", report.Subject.Namespace, edgeOrdinal),
				Kind: item.program.kind, Label: item.program.label, FromActivity: producer.Name, ToActivity: activity.Name, ViaEntity: input.Name,
			})
		}
	}
	report.Summary.TypedEdges = len(report.Edges)
	report.Summary.UnresolvedInputs = len(report.Gaps)
	countClaimDependencyKinds(&report)
	if len(report.Gaps) > 0 {
		blocked := make([]string, 0, len(report.Gaps))
		for _, gap := range report.Gaps {
			blocked = append(blocked, "entity:"+gap.InputEntity)
		}
		return unknownClaimDependencies(report, "DEPENDENCY_DISCOVERY", "BIND_INPUT_PRODUCER", "CLAIM_INPUT_PRODUCER_UNAVAILABLE", "DIRECT_MISSING", "DECLARE_INPUT_PRODUCER", claimDependencyCoherence, blocked)
	}
	cycle := claimDependencyCycleResidue(report.Nodes, report.Edges)
	report.Summary.CyclicActivities = len(cycle)
	if len(cycle) > 0 {
		return refuteClaimDependencies(report, "CAUSALITY", "SELECT_PROOF_STRUCTURE", "CLAIM_DEPENDENCY_CYCLE_DETECTED", "SELECT_FOUNDATION_OR_BREAK_CYCLE", claimDependencyRegression, cycle)
	}
	if report.Summary.RecoverableRoots == 0 {
		return unknownClaimDependencies(report, "FOUNDATION", "SELECT_RECOVERABLE_ROOT", "RECOVERABLE_ROOT_UNAVAILABLE", "DIRECT_MISSING", "DECLARE_RECOVERABLE_ROOT", claimDependencyFoundation, []string{})
	}
	report.Decision = claimDependencyObserved
	report.Resolution = claimDependencyResolution{
		State: claimStateClosed, Reason: "CLAIM_DEPENDENCY_CAUSALITY_OBSERVED", NextOperation: claimResolutionNone,
		ProofChoice: claimDependencyCoherence, BlockedBy: []string{},
	}
	report.Indicators = buildClaimDependencyIndicators(report)
	return report
}

func newClaimDependencyReport(sourceFile string, source []byte, file *syntax.File) claimDependencyReport {
	namespace := ""
	if file.Namespace != nil {
		namespace = file.Namespace.Name
	}
	return claimDependencyReport{
		Schema: claimDependencySchema, Candidate: claimDependencyCandidateID, Decision: claimDependencyFailed,
		Subject: claimDependencySubject{SourceFile: sourceFile, SourceDigest: claimResolutionDigest(source), Namespace: namespace, Binding: "GOOO_ACTIVITY_DATAFLOW_AND_VALUE_PROGRAM"},
		Contract: claimDependencyContract{
			Version: "v1", RootProgram: claimDependencyRootProgram, EdgeProgramPrefix: claimDependencyEdgePrefix,
			EdgeKinds: append([]string{}, claimDependencyEdgeKinds...), GraphSemantics: "STRUCTURAL_ONLY",
			CyclePolicy: "FAIL_CLOSED", MissingInputPolicy: "LOWER_TO_UNKNOWN",
		},
		Nodes: []claimDependencyNode{}, Edges: []claimDependencyEdge{}, Gaps: []claimDependencyGap{}, Indicators: []claimDependencyIndicator{},
		Authority: claimDependencyAuthority{Source: "GOOO_SOURCE", RepositoryWrites: 0},
	}
}

func unknownClaimDependencies(report claimDependencyReport, stage, step, reason, unknownClass, next, proof string, blockedBy []string) claimDependencyReport {
	report.Decision = claimDependencyIncomplete
	report.Resolution = claimDependencyResolution{
		State: claimStateUnknown, Stage: claimOptional(stage), Step: claimOptional(step), Reason: reason,
		UnknownClass: claimOptional(unknownClass), NextOperation: next, ProofChoice: proof, BlockedBy: blockedBy,
	}
	report.Indicators = buildClaimDependencyIndicators(report)
	return report
}

func refuteClaimDependencies(report claimDependencyReport, stage, step, reason, next, proof string, blockedBy []string) claimDependencyReport {
	report.Decision = claimDependencyFailed
	report.Resolution = claimDependencyResolution{
		State: claimStateRefuted, Stage: claimOptional(stage), Step: claimOptional(step), Reason: reason,
		NextOperation: next, ProofChoice: proof, BlockedBy: blockedBy,
	}
	report.Indicators = buildClaimDependencyIndicators(report)
	return report
}

func countClaimDependencyKinds(report *claimDependencyReport) {
	for _, edge := range report.Edges {
		switch edge.Kind {
		case "REQUIRES":
			report.KindCounts.Requires++
		case "SUPPORTS":
			report.KindCounts.Supports++
		case "CONTRADICTS":
			report.KindCounts.Contradicts++
		case "FAILURE_ENTAILMENT":
			report.KindCounts.FailureEntailment++
		}
	}
}
