package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func loadMetricSubjects(name, expectedSHA string) ([]inputSubject, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var report linecaps.LineMetricsReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	if report.CommitSHA != expectedSHA {
		return nil, fmt.Errorf("metrics SHA %s does not match %s", report.CommitSHA, expectedSHA)
	}
	selected := make([]inputSubject, 0)
	seen := make(map[string]bool)
	for _, indicator := range report.Meta.Indicators {
		if indicator.MetricID != sourcepolicy.DimensionGoFileLines ||
			indicator.Applicability == sourcepolicy.ApplicabilityNotApplicable ||
			indicator.Satisfied || !indicator.Blocking ||
			indicator.Operation != sourcepolicy.OperationSplitGo {
			continue
		}
		if indicator.Subject == "" || seen[indicator.Subject] {
			if indicator.Subject == "" {
				return nil, fmt.Errorf("metrics contain an empty source subject")
			}
			return nil, fmt.Errorf("duplicate metric subject %s", indicator.Subject)
		}
		seen[indicator.Subject] = true
		selected = append(selected, inputSubject{
			Indicator: string(indicator.MetricID), Logical: indicator.Subject,
			Value: indicator.Value, Limit: indicator.Limit,
			Consumer: string(indicator.Consumer), Operation: string(indicator.Operation),
		})
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Logical < selected[j].Logical })
	return selected, nil
}
