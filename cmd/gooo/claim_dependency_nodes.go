package main

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

func observeClaimDependencyPrograms(report claimDependencyReport, activities []*syntax.ActivityDecl) (claimDependencyReport, []parsedClaimDependencyActivity, bool) {
	parsed := make([]parsedClaimDependencyActivity, 0, len(activities))
	kinds := make(map[string]bool, len(claimDependencyEdgeKinds))
	for index, activity := range activities {
		if !activity.ValueProgramPresent {
			report = refuteClaimDependencies(report, "CLAIM_DEPENDENCY", "OBSERVE_VALUE_PROGRAM", "CLAIM_DEPENDENCY_PROGRAM_MISSING", "DECLARE_CLAIM_DEPENDENCY_VALUE_PROGRAM", claimDependencyFoundation, []string{activity.Name})
			return report, nil, false
		}
		program, failed := parseClaimDependencyProgram(activity.ValueProgram)
		if failed != nil {
			report = refuteClaimDependencies(report, "CLAIM_DEPENDENCY", "PARSE_VALUE_PROGRAM", failed.reason, failed.next, claimDependencyRegression, []string{activity.Name})
			return report, nil, false
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
	return report, parsed, true
}
