package main

import (
	"fmt"
	"sort"
	"strings"
)

const (
	functionMetric      = "gooo.metric.refactor.single-return.v1"
	functionLinesMetric = "gooo.metric.source.function-lines.v1"
)

func compileFunctionWitnesses(indicators []sourceIndicator) ([]subjectWitness, error) {
	bySubject := make(map[string][]sourceIndicator)
	for _, row := range indicators {
		if row.SubjectKind != "FUNCTION" {
			continue
		}
		switch row.MetricID {
		case functionMetric:
			if err := validateFunctionIndicator(row); err != nil {
				return nil, err
			}
		case functionLinesMetric:
			if err := validateLineCapIndicator(row); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("function indicator %q is outside the exact catalog", row.MetricID)
		}
		rows := bySubject[row.Subject]
		for _, existing := range rows {
			if existing.MetricID == row.MetricID {
				return nil, fmt.Errorf("function subject %q metric %q is duplicated", row.Subject, row.MetricID)
			}
		}
		bySubject[row.Subject] = append(rows, row)
	}

	subjects := make([]string, 0, len(bySubject))
	for subject := range bySubject {
		subjects = append(subjects, subject)
	}
	sort.Strings(subjects)
	witnesses := make([]subjectWitness, 0, len(subjects))
	for _, subject := range subjects {
		rows := bySubject[subject]
		sort.Slice(rows, func(i, j int) bool { return rows[i].MetricID < rows[j].MetricID })
		metrics := make([]metricValue, 0, len(rows))
		for _, row := range rows {
			metrics = append(metrics, metricValue{ID: row.MetricID, Value: row.Value})
		}
		witness := subjectWitness{Space: "LOGICAL_FUNCTION", Path: subject, SubjectKind: "FUNCTION", Metrics: metrics, Meta: sourceBinding(rows, "COHERENCE")}
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
