package main

import (
	"fmt"
	"strings"
)

const functionMetric = "gooo.metric.refactor.single-return.v1"

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
