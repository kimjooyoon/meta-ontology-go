package main

import (
	"encoding/json"
	"fmt"
)

func sourceReadme(indicators []sourceIndicatorEvidence) (sourceReadmeObservation, error) {
	const metricID = "gooo.metric.documentation.root-readme-presence.v1"
	var matches []sourceIndicatorEvidence
	for _, indicator := range indicators {
		if indicator.MetricID == metricID {
			matches = append(matches, indicator)
		}
	}
	if len(matches) != 1 {
		return sourceReadmeObservation{}, fmt.Errorf("source metrics root README indicator is not unique")
	}
	indicator := matches[0]
	return sourceReadmeObservation{Applicability: indicator.Applicability, Reason: indicator.ApplicabilityReason, Blocking: indicator.Blocking}, nil
}

func sourceCandidates(indicators []sourceIndicatorEvidence) []sourceCandidate {
	candidates := make([]sourceCandidate, 0)
	for _, indicator := range indicators {
		if !sourceCandidateMetric(indicator.MetricID) || indicator.Applicability != "APPLICABLE" || indicator.Role != "DRIVER" || indicator.Value <= indicator.Limit {
			continue
		}
		candidates = append(candidates, sourceCandidate{MetricID: indicator.MetricID, Subject: indicator.Subject, Actual: indicator.Value, Threshold: indicator.Limit, Role: "DRIVER"})
	}
	return candidates
}

func sourceCandidateMetric(metricID string) bool {
	return metricID == "gooo.metric.source.go-file-lines.v1" || metricID == "gooo.metric.source.gooo-file-lines.v1" || metricID == "gooo.metric.source.function-lines.v1"
}

func decodeSelectedSubjects(data []byte) ([]selectedSubject, error) {
	var plan planSelectionEvidence
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("decode self-improvement plan: %w", err)
	}
	if plan.SchemaVersion != "gooo/self-improvement-generation/v6" {
		return nil, fmt.Errorf("self-improvement plan schema is not v6")
	}
	for index := range plan.Selected {
		if plan.Selected[index].MetaOperation == "" || plan.Selected[index].MetricID == "" || plan.Selected[index].Subject == "" {
			return nil, fmt.Errorf("selected plan subject %d is incomplete", index)
		}
	}
	return plan.Selected, nil
}
