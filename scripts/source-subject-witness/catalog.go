package main

const defaultApplicabilityRule = "gooo.catalog.source-policy.default-applicability.v1"

func exactApplicable(row sourceIndicator) bool {
	return row.Applicability == "APPLICABLE" && row.ApplicabilityRuleID == defaultApplicabilityRule && row.ApplicabilityReason == "CATALOG_APPLICABLE" && row.Satisfied && row.Decision == "PASS" && row.EvaluationState == "EVALUATED" && row.FailureReason == "NONE" && row.EnforcementEffect != ""
}
