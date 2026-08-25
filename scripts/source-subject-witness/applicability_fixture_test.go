package main

func workflowDiscoveryFixture() (directoryMetric, []sourceIndicator) {
	directory := directoryMetric{Path: workflowDiscoveryPath, SubjectKind: "DIRECTORY", DirectFiles: 16, RecursiveFiles: 16}
	metrics := []expectedMetric{
		{"gooo.metric.layout.direct-entries.v1", 16},
		{"gooo.metric.layout.direct-files.v1", 16},
		{"gooo.metric.layout.direct-folders.v1", 0},
		{"gooo.metric.layout.entry-kinds.v1", 1},
		{"gooo.metric.layout.recursive-files.v1", 16},
		{"gooo.metric.layout.recursive-folders.v1", 0},
	}
	rows := make([]sourceIndicator, 0, len(metrics))
	for _, metric := range metrics {
		rows = append(rows, sourceIndicator{Subject: directory.Path, SubjectKind: directory.SubjectKind, MetricID: metric.id, Value: metric.value, Applicability: "APPLICABLE", Decision: "PASS", Satisfied: true, MetaOperation: "observe"})
	}
	rows[0].Applicability = "NOT_APPLICABLE"
	rows[0].ApplicabilityRuleID = workflowDiscoveryRule
	rows[0].ApplicabilityReason = "GITHUB_WORKFLOW_DISCOVERY_ROOT"
	rows[0].MetaOperation = workflowDiscoveryOperation
	rows[0].ProofChoice = "foundation"
	rows[0].Consumer = "github-actions"
	rows[0].Decision = "NOT_APPLICABLE"
	rows[0].EvaluationState = "EVALUATED"
	rows[0].FailureReason = "CATALOG_NOT_APPLICABLE"
	rows[0].EnforcementEffect = "NO_EFFECT"
	return directory, rows
}
