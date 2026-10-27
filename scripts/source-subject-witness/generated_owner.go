package main

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// Decode the shared wire contract instead of copying the source-owned rule.
// The observation state and producer still belong to this native consumer.
func generatedOwnerIndicator(row sourceIndicator) bool {
	contract, known := lineCapContract(row.MetricID)
	if !known || row.Applicability != "NOT_APPLICABLE" {
		return false
	}
	raw, err := json.Marshal(row)
	var indicator sourcepolicy.Indicator
	if err != nil || json.Unmarshal(raw, &indicator) != nil {
		return false
	}
	return sourcepolicy.ValidGeneratedOwnershipShape(indicator) && row.Producer == contract.producer &&
		row.Satisfied && !row.Blocking && row.Decision == "NOT_APPLICABLE" &&
		row.EvaluationState == "EVALUATED" && row.EnforcementEffect == "NO_EFFECT" &&
		row.FailureReason == "CATALOG_NOT_APPLICABLE" && row.FailureCode == row.MetricID+"#catalog-not-applicable"
}
