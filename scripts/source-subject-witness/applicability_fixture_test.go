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
	for _, index := range []int{0, 3} {
		rows[index].Applicability = "NOT_APPLICABLE"
		rows[index].ApplicabilityRuleID = workflowDiscoveryRule
		rows[index].ApplicabilityReason = "WORKFLOW_DISCOVERY_ROOT_EXEMPT"
		rows[index].MetaOperation = workflowDiscoveryOperation
		rows[index].ProofChoice = "foundation"
		rows[index].Consumer = "github-actions"
		rows[index].Decision = "NOT_APPLICABLE"
		rows[index].EvaluationState = "EVALUATED"
		rows[index].FailureReason = "CATALOG_NOT_APPLICABLE"
		rows[index].EnforcementEffect = "NO_EFFECT"
	}
	return directory, rows
}
