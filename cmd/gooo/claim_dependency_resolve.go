package main

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

type parsedClaimDependencyActivity struct {
	declaration *syntax.ActivityDecl
	program     claimDependencyProgram
}

func resolveClaimDependencies(sourceFile string, source []byte, file *syntax.File) claimDependencyReport {
	report := newClaimDependencyReport(sourceFile, source, file)
	report, activities, producers, ok := collectClaimDependencyActivities(report, file)
	if !ok {
		return report
	}
	report, parsed, ok := observeClaimDependencyPrograms(report, activities)
	if !ok {
		return report
	}
	report, ok = bindClaimDependencyEdges(report, parsed, producers)
	if !ok {
		return report
	}
	return finalizeClaimDependencies(report)
}
