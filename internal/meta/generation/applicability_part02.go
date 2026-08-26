package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func validActionApplicability(action Action) bool {
	return action.SubjectKind != "" &&
		action.Applicability == sourcepolicy.ApplicabilityApplicable &&
		action.ApplicabilityRule == sourcepolicy.ApplicabilityRuleDefault &&
		action.ApplicabilityReason == sourcepolicy.ApplicabilityReasonCatalogApplicable &&
		validMetricProof(action.MetricProofChoice) &&
		action.MetricProducer != "" && action.MetricConsumer != ""
}

func planApplicabilityKnown(plan Plan, selected map[string]Action) bool {
	if plan.NotApplicableIndicatorIDs == nil ||
		!validOrderedIndicatorIDs(plan.NotApplicableIndicatorIDs) {
		return false
	}
	for _, identifier := range plan.NotApplicableIndicatorIDs {
		if _, executable := selected[identifier]; executable {
			return false
		}
	}
	return true
}

func validOrderedIndicatorIDs(identifiers []string) bool {
	for index, identifier := range identifiers {
		if !validActionIndicatorID(identifier) ||
			(index > 0 && identifiers[index-1] >= identifier) {
			return false
		}
	}
	return true
}
