package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func knownModuleBoundaryReason(reason string) bool {
	switch sourcepolicy.ApplicabilityReason(reason) {
	case sourcepolicy.ApplicabilityReasonModuleOwnerUnavailable,
		sourcepolicy.ApplicabilityReasonCrossModuleSource,
		sourcepolicy.ApplicabilityReasonModuleToolchainUnavailable:
		return true
	}
	return false
}

func validateRecordedModuleBoundary(file fileMetric, row sourceIndicator) error {
	if file.ModuleBoundaryReason == "" && row.ApplicabilityRuleID != sourcepolicy.ApplicabilityRuleGoModuleBoundary {
		return nil
	}
	if file.Language != "go" || !knownModuleBoundaryReason(file.ModuleBoundaryReason) ||
		row.Applicability != "NOT_APPLICABLE" || row.ApplicabilityRuleID != sourcepolicy.ApplicabilityRuleGoModuleBoundary ||
		row.ApplicabilityReason != file.ModuleBoundaryReason {
		return sourceValidationFailure("SOURCE_MODULE_BOUNDARY_CONTRADICTION", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	return nil
}

func validateModuleLineCap(row sourceIndicator, contract lineCapSpec, known bool) error {
	valid := known && row.MetricID != "gooo.metric.source.gooo-file-lines.v1" &&
		row.SubjectKind == contract.kind && row.Family == contract.family &&
		row.ApplicabilityRuleID == sourcepolicy.ApplicabilityRuleGoModuleBoundary &&
		knownModuleBoundaryReason(row.ApplicabilityReason) &&
		row.Relation == "less_or_equal" && row.Role == "DRIVER" && !row.Blocking &&
		row.Producer == contract.producer && row.Consumer == contract.consumer && row.MetaOperation == contract.operation &&
		row.ProofChoice == "foundation" && row.Decision == "NOT_APPLICABLE" && row.Satisfied &&
		row.EvaluationState == "EVALUATED" && row.EnforcementEffect == "NO_EFFECT" &&
		row.FailureReason == "CATALOG_NOT_APPLICABLE" && row.FailureCode == row.MetricID+"#catalog-not-applicable"
	if !valid {
		return sourceValidationFailure("SOURCE_MODULE_LINE_CAP_CONTRADICTION", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	return nil
}
