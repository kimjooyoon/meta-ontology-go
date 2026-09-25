package main

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

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
