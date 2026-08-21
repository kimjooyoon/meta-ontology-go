package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func applicabilityFailures(indicators []sourcepolicy.Indicator) []string {
	failures := make([]string, 0)
	for _, indicator := range indicators {
		if !validIndicatorApplicability(indicator) {
			failures = append(failures, indicatorID(indicator))
		}
	}
	return failures
}

func notApplicableIndicatorIDs(indicators []sourcepolicy.Indicator) []string {
	result := make([]string, 0)
	for _, indicator := range indicators {
		if indicator.Applicability == sourcepolicy.ApplicabilityNotApplicable {
			result = append(result, indicatorID(indicator))
		}
	}
	return result
}

func validIndicatorApplicability(indicator sourcepolicy.Indicator) bool {
	if indicator.MetricID == "" || indicator.Subject == "" || indicator.SubjectKind == "" ||
		indicator.ApplicabilityRule == "" || !validMetricProof(indicator.Proof) ||
		indicator.Producer == "" || indicator.Consumer == "" || indicator.Operation == "" {
		return false
	}
	switch indicator.Applicability {
	case sourcepolicy.ApplicabilityApplicable:
		return indicator.ApplicabilityRule == sourcepolicy.ApplicabilityRuleDefault &&
			indicator.ApplicabilityReason == sourcepolicy.ApplicabilityReasonCatalogApplicable
	case sourcepolicy.ApplicabilityNotApplicable:
		return indicator.Subject == "." &&
			indicator.SubjectKind == sourcepolicy.SubjectKindProjectRoot &&
			indicator.ApplicabilityRule == sourcepolicy.ApplicabilityRuleProjectRootTopology &&
			indicator.ApplicabilityReason == sourcepolicy.ApplicabilityReasonRootTopologyExempt &&
			indicator.Operation == sourcepolicy.OperationExemptRoot &&
			indicator.Satisfied && !indicator.Blocking
	default:
		return false
	}
}

func validMetricProof(proof sourcepolicy.ProofChoice) bool {
	return proof == sourcepolicy.ProofFoundation ||
		proof == sourcepolicy.ProofCoherence ||
		proof == sourcepolicy.ProofRegression
}
