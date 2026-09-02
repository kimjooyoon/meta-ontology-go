package main

import "fmt"

const (
	workflowDiscoveryPath      = ".github/workflows"
	workflowDiscoveryMetric    = "gooo.metric.layout.direct-entries.v1"
	workflowDiscoveryRule      = "gooo.catalog.source-policy.github-workflow-discovery.v1"
	workflowDiscoveryOperation = "preserve-workflow-discovery"
)

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
	if len(exemptions) != 1 || !isWorkflowDiscoveryExemption(exemptions[0]) {
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
	return row.MetricID == workflowDiscoveryMetric && row.ApplicabilityRuleID == workflowDiscoveryRule &&
		row.ApplicabilityReason == "GITHUB_WORKFLOW_DISCOVERY_ROOT" && row.MetaOperation == workflowDiscoveryOperation &&
		row.ProofChoice == "foundation" && row.Consumer == "github-actions" && row.Decision == "NOT_APPLICABLE" &&
		row.EvaluationState == "EVALUATED" && row.FailureReason == "CATALOG_NOT_APPLICABLE" &&
		row.EnforcementEffect == "NO_EFFECT" && !row.Blocking
}
