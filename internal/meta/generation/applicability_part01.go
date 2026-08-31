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
		if indicator.ApplicabilityRule != sourcepolicy.ApplicabilityRuleDefault ||
			indicator.ApplicabilityReason != sourcepolicy.ApplicabilityReasonCatalogApplicable {
			return false
		}
		if sourcepolicy.IsLineCapMetric(indicator.MetricID) {
			return indicator.Role == sourcepolicy.IndicatorRoleDriver && !indicator.Blocking &&
				indicator.Relation == sourcepolicy.RelationLessOrEqual &&
				indicator.Satisfied == (indicator.Value <= indicator.Limit)
		}
		return true
	case sourcepolicy.ApplicabilityNotApplicable:
		if !indicator.Satisfied || indicator.Blocking {
			return false
		}
		if indicator.Operation == sourcepolicy.OperationPreserveWorkflow {
			return indicator.Subject == ".github/workflows" && indicator.SubjectKind == sourcepolicy.SubjectKindDirectory &&
				indicator.ApplicabilityRule == sourcepolicy.ApplicabilityRuleWorkflowDiscovery &&
				indicator.ApplicabilityReason == sourcepolicy.ApplicabilityReasonWorkflowDiscovery
		}
		if indicator.Subject != "." || indicator.SubjectKind != sourcepolicy.SubjectKindProjectRoot {
			return false
		}
		switch indicator.Operation {
		case sourcepolicy.OperationExemptWorkflowRoot:
			return indicator.Subject == ".github/workflows" &&
				indicator.SubjectKind == sourcepolicy.SubjectKindDirectory &&
				indicator.ApplicabilityRule == sourcepolicy.ApplicabilityRuleWorkflowDiscoveryRoot &&
				indicator.ApplicabilityReason == sourcepolicy.ApplicabilityReasonWorkflowRootExempt
		case sourcepolicy.OperationExemptRoot:
			if indicator.Subject != "." || indicator.SubjectKind != sourcepolicy.SubjectKindProjectRoot {
				return false
			}
			return indicator.ApplicabilityRule == sourcepolicy.ApplicabilityRuleProjectRootTopology &&
				indicator.ApplicabilityReason == sourcepolicy.ApplicabilityReasonRootTopologyExempt
		case sourcepolicy.OperationExemptRootREADME:
			if indicator.Subject != "." || indicator.SubjectKind != sourcepolicy.SubjectKindProjectRoot {
				return false
			}
			return indicator.ApplicabilityRule == sourcepolicy.ApplicabilityRuleProjectRootREADME &&
				indicator.ApplicabilityReason == sourcepolicy.ApplicabilityReasonRootREADMEExempt
		default:
			return false
		}
	default:
		return false
	}
}

func validMetricProof(proof sourcepolicy.ProofChoice) bool {
	return proof == sourcepolicy.ProofFoundation ||
		proof == sourcepolicy.ProofCoherence ||
		proof == sourcepolicy.ProofRegression
}
