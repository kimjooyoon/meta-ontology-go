package languageassurance

import "fmt"

func detectSelfMinting(transaction Transaction) []Finding {
	var findings []Finding
	for _, route := range transaction.AuthorityRoutes {
		if route.AuthoredBy == route.PromotedBy {
			findings = append(findings, Finding{MetricID: MetricSelfMinting, PathID: "authority/" + route.RuleID + "/" + route.AuthoredBy, Principal: route.AuthoredBy, RuleID: route.RuleID})
		}
	}
	return findings
}

func detectRoleConflicts(transaction Transaction) []Finding {
	var findings []Finding
	for _, binding := range transaction.RoleBindings {
		roles := roleSet(binding.Roles)
		for _, pair := range conflictPairs {
			if roles[pair.Left] && roles[pair.Right] {
				path := fmt.Sprintf("role/%s/%s+%s", binding.Principal, pair.Left, pair.Right)
				findings = append(findings, Finding{MetricID: MetricRoleConflict, PathID: path, Principal: binding.Principal, Roles: []Role{pair.Left, pair.Right}})
			}
		}
	}
	return findings
}

func detectUnknownLaundering(transaction Transaction) []Finding {
	var findings []Finding
	for _, transition := range transaction.DecisionTransitions {
		if transition.Input == DecisionUnknown && isLaunderingOutput(transition.Output) {
			path := fmt.Sprintf("decision/%s/%s->%s", transition.ID, transition.Input, transition.Output)
			findings = append(findings, Finding{MetricID: MetricUnknownLaundering, PathID: path, DecisionID: transition.ID, Input: transition.Input, Output: transition.Output})
		}
	}
	return findings
}

func countFindings(findings []Finding, metricID string) int {
	count := 0
	for _, finding := range findings {
		if finding.MetricID == metricID {
			count++
		}
	}
	return count
}

func countUnknownTop(transitions []DecisionTransition) int {
	count := 0
	for _, transition := range transitions {
		if transition.Input == DecisionUnknown {
			count++
		}
	}
	return count
}

func observedValue(observed bool, value int) *int {
	if !observed {
		return nil
	}
	return &value
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
