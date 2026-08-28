package main

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

func collectClaimDependencyActivities(report claimDependencyReport, file *syntax.File) (claimDependencyReport, []*syntax.ActivityDecl, map[string]*syntax.ActivityDecl, bool) {
	activities := make([]*syntax.ActivityDecl, 0)
	producers := make(map[string]*syntax.ActivityDecl)
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		report.Summary.ActivitiesTotal++
		if prior, duplicate := producers[activity.Output]; duplicate {
			report = refuteClaimDependencies(report, "META_BINDING", "BIND_OUTPUT_PRODUCERS", "CLAIM_OUTPUT_PRODUCER_AMBIGUOUS", "RESTORE_ONE_PRODUCER_PER_ENTITY", claimDependencyRegression, []string{prior.Name, activity.Name})
			return report, nil, nil, false
		}
		producers[activity.Output] = activity
		activities = append(activities, activity)
	}
	if len(activities) == 0 {
		report = unknownClaimDependencies(report, "CLAIM_DEPENDENCY", "OBSERVE_CLAIM_ACTIVITIES", "CLAIM_DEPENDENCY_ACTIVITIES_UNAVAILABLE", "DIRECT_MISSING", "DECLARE_CLAIM_DEPENDENCY_ACTIVITIES", claimDependencyFoundation, []string{})
		return report, nil, nil, false
	}
	return report, activities, producers, true
}
