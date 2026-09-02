package main

import "fmt"

const (
	workflowDiscoveryPath          = ".github/workflows"
	workflowDiscoveryEntriesMetric = "gooo.metric.layout.direct-entries.v1"
	workflowDiscoveryKindsMetric   = "gooo.metric.layout.entry-kinds.v1"
	workflowDiscoveryRule          = "gooo.catalog.source-policy.workflow-discovery-root.v1"
	workflowDiscoveryOperation     = "exempt-workflow-discovery-root"
)

const workflowDiscoveryExemptionRows = 2

func validateDirectoryApplicability(directory directoryMetric, rows []sourceIndicator) error {
	exemptions := notApplicableRows(rows)
	if directory.Path == "." {
		if len(exemptions) != 3 {
			return fmt.Errorf("directory %q has %d topology exemptions, want 3", directory.Path, len(exemptions))
		}
		return nil
	}
	if directory.Path != workflowDiscoveryPath {
		if len(exemptions) != 0 {
			return fmt.Errorf("directory %q has %d unknown topology exemptions", directory.Path, len(exemptions))
		}
		return nil
	}
	if len(exemptions) != workflowDiscoveryExemptionRows {
		return fmt.Errorf("directory %q does not have the exact workflow discovery exemption", directory.Path)
	}
	seen := make(map[string]bool, len(exemptions))
	for _, row := range exemptions {
		if !isWorkflowDiscoveryExemption(row) || seen[row.MetricID] {
			return fmt.Errorf("directory %q does not have the exact workflow discovery exemption", directory.Path)
		}
		seen[row.MetricID] = true
	}
	if !seen[workflowDiscoveryEntriesMetric] || !seen[workflowDiscoveryKindsMetric] {
		return fmt.Errorf("directory %q does not have the exact workflow discovery exemption", directory.Path)
	}
	return nil
}

func notApplicableRows(rows []sourceIndicator) []sourceIndicator {
	values := make([]sourceIndicator, 0)
	for _, row := range rows {
		if row.Applicability == "NOT_APPLICABLE" {
			values = append(values, row)
		}
	}
	return values
}

func isWorkflowDiscoveryExemption(row sourceIndicator) bool {
	return (row.MetricID == workflowDiscoveryEntriesMetric || row.MetricID == workflowDiscoveryKindsMetric) && row.ApplicabilityRuleID == workflowDiscoveryRule &&
		row.ApplicabilityReason == "WORKFLOW_DISCOVERY_ROOT_EXEMPT" && row.MetaOperation == workflowDiscoveryOperation &&
		row.ProofChoice == "foundation" && row.Consumer == "github-actions" && row.Decision == "NOT_APPLICABLE" &&
		row.EvaluationState == "EVALUATED" && row.FailureReason == "CATALOG_NOT_APPLICABLE" &&
		row.EnforcementEffect == "NO_EFFECT" && !row.Blocking
}
