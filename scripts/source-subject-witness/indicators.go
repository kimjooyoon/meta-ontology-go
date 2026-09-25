package main

import "fmt"

type sourceIndicator struct {
	Applicability       string `json:"applicability"`
	ApplicabilityReason string `json:"applicability_reason"`
	ApplicabilityRuleID string `json:"applicability_rule_id"`
	Blocking            bool   `json:"blocking"`
	Consumer            string `json:"consumer"`
	Decision            string `json:"decision"`
	Detail              string `json:"detail"`
	EnforcementEffect   string `json:"enforcement_effect"`
	EvaluationState     string `json:"evaluation_state"`
	FailureCode         string `json:"failure_code"`
	FailureReason       string `json:"failure_reason"`
	Family              string `json:"family"`
	Limit               int    `json:"limit"`
	MetaOperation       string `json:"meta_operation"`
	MetricID            string `json:"metric_id"`
	Producer            string `json:"producer"`
	ProofChoice         string `json:"proof_choice"`
	Relation            string `json:"relation"`
	Role                string `json:"role"`
	Satisfied           bool   `json:"satisfied"`
	Subject             string `json:"subject"`
	SubjectKind         string `json:"subject_kind"`
	Value               int    `json:"value"`
}

type indicatorIndex map[string][]sourceIndicator

func indexIndicators(values []sourceIndicator) indicatorIndex {
	index := make(indicatorIndex)
	for _, value := range values {
		key := indicatorKey(value.SubjectKind, value.Subject, value.MetricID)
		index[key] = append(index[key], value)
	}
	return index
}

func indicatorKey(kind, subject, metric string) string {
	return kind + "\x00" + subject + "\x00" + metric
}

func lookupIndicator(index indicatorIndex, kind, subject, metric string, want int) (sourceIndicator, error) {
	rows := index[indicatorKey(kind, subject, metric)]
	if len(rows) != 1 {
		return sourceIndicator{}, fmt.Errorf("%s %q %s has %d indicators", kind, subject, metric, len(rows))
	}
	row := rows[0]
	if row.Value != want {
		return sourceIndicator{}, fmt.Errorf("%s %q %s observed %d, want %d", kind, subject, metric, row.Value, want)
	}
	return row, nil
}

type ledgerIndicator struct {
	ID       string `json:"id"`
	Route    string `json:"route"`
	Verdict  string `json:"verdict"`
	Relation string `json:"relation"`
	Value    string `json:"value"`
	Limit    string `json:"limit"`
}

func nonNegative(values ...int) bool {
	for _, value := range values {
		if value < 0 {
			return false
		}
	}
	return true
}
