package main

import (
	"fmt"
	"strings"
)

const (
	defaultApplicabilityRule = "gooo.catalog.source-policy.default-applicability.v1"
	functionMetric           = "gooo.metric.refactor.single-return.v1"
	rootUnboundMetric        = "gooo.metric.meta.unbound-indicators.v1"
	rootSummaryCount         = 5
)

func rootSummaryMetrics(directory directoryMetric) []expectedMetric {
	return []expectedMetric{
		{"gooo.metric.source.go-files.v1", directory.GoFiles},
		{"gooo.metric.source.go-lines.v1", directory.GoLines},
		{"gooo.metric.source.gooo-files.v1", directory.GoooFiles},
		{"gooo.metric.source.gooo-lines.v1", directory.GoooLines},
		{rootUnboundMetric, 0},
	}
}

func validateRootSummaryIndicator(row sourceIndicator) error {
	if !isRootSummaryMetric(row.MetricID) {
		return nil
	}
	wantOperation, wantProducer, wantConsumer, wantProof, wantBlocking := "observe", "linecaps.AnalyzeLineMetrics", "metric-report", "coherence", false
	if row.MetricID == rootUnboundMetric {
		wantOperation, wantProducer, wantConsumer, wantBlocking = "bind-indicator-meta-program", "metabinding.Build", "self-improvement-cycle", true
		if row.Value != 0 || row.Relation != "less_or_equal" || row.Detail != "ontology=examples/meta-binding-coverage/main.gooo" {
			return fmt.Errorf("root unbound indicator lost its exact zero-value ontology binding")
		}
	}
	if row.Subject != "." || row.SubjectKind != "PROJECT_ROOT" || row.MetaOperation != wantOperation || row.Producer != wantProducer || row.Consumer != wantConsumer || row.ProofChoice != wantProof || row.Blocking != wantBlocking || !exactApplicable(row) {
		return fmt.Errorf("root summary indicator %q is outside the exact catalog", row.MetricID)
	}
	return nil
}

func isRootSummaryMetric(metric string) bool {
	for _, item := range rootSummaryMetrics(directoryMetric{}) {
		if item.id == metric {
			return true
		}
	}
	return false
}

func compileFunctionWitnesses(indicators []sourceIndicator) ([]subjectWitness, error) {
	witnesses, seen := make([]subjectWitness, 0), make(map[string]bool)
	for _, row := range indicators {
		if row.SubjectKind != "FUNCTION" {
			continue
		}
		if err := validateFunctionIndicator(row); err != nil {
			return nil, err
		}
		if seen[row.Subject] {
			return nil, fmt.Errorf("function subject %q is duplicated", row.Subject)
		}
		seen[row.Subject] = true
		witness := subjectWitness{Space: "LOGICAL_FUNCTION", Path: row.Subject, SubjectKind: row.SubjectKind, Metrics: []metricValue{{ID: row.MetricID, Value: row.Value}}, Meta: sourceBinding([]sourceIndicator{row}, "COHERENCE")}
		witnesses = append(witnesses, sealWitness(witness))
	}
	return witnesses, nil
}

func validateFunctionIndicator(row sourceIndicator) error {
	exactCatalog := row.MetricID == functionMetric && row.Value == 1 && row.Limit == 0 && row.Relation == "observe" && row.MetaOperation == "inspect-wrapper" && row.Producer == "linecaps.Analyze" && row.Consumer == "refactor-report" && row.ProofChoice == "coherence" && !row.Blocking && strings.HasPrefix(row.Detail, "single return ")
	if row.Subject == "" || !exactCatalog || !exactApplicable(row) {
		return fmt.Errorf("function indicator %q is outside the exact single-return catalog", row.Subject)
	}
	return nil
}

func exactApplicable(row sourceIndicator) bool {
	return row.Applicability == "APPLICABLE" && row.ApplicabilityRuleID == defaultApplicabilityRule && row.ApplicabilityReason == "CATALOG_APPLICABLE" && row.Satisfied && row.Decision == "PASS" && row.EvaluationState == "EVALUATED" && row.FailureReason == "NONE" && row.EnforcementEffect != ""
}
