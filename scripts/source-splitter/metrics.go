package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type metricReport struct {
	CommitSHA   string            `json:"commit_sha"`
	Directories []metricDirectory `json:"directories"`
	Meta        struct {
		Policy struct {
			MaxFileLines     int `json:"max_file_lines"`
			MaxDirectEntries int `json:"max_direct_entries"`
		} `json:"policy"`
		Indicators []metricIndicator `json:"indicators"`
	} `json:"meta"`
}

type metricDirectory struct {
	Path          string `json:"path"`
	DirectFolders int    `json:"direct_folders"`
	DirectFiles   int    `json:"direct_files"`
}

type metricIndicator struct {
	MetricID     sourcepolicy.Dimension `json:"metric_id"`
	Subject      string                 `json:"subject"`
	Satisfied    bool                   `json:"satisfied"`
	MetaOperation sourcepolicy.Operation `json:"meta_operation"`
}

func loadMetricReport(path, expectedSHA string) (metricReport, error) {
	var report metricReport
	data, err := os.ReadFile(path)
	if err != nil {
		return report, fmt.Errorf("read metrics: %w", err)
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return report, fmt.Errorf("decode metrics: %w", err)
	}
	if report.CommitSHA != expectedSHA {
		return report, fmt.Errorf("metrics SHA %s does not match %s", report.CommitSHA, expectedSHA)
	}
	if report.Meta.Policy.MaxFileLines <= 0 || report.Meta.Policy.MaxDirectEntries <= 0 {
		return report, fmt.Errorf("metrics policy limits are missing")
	}
	return report, nil
}

func (report metricReport) splitIndicators() []metricIndicator {
	indicators := make([]metricIndicator, 0)
	for _, indicator := range report.Meta.Indicators {
		if !indicator.Satisfied && indicator.MetricID == sourcepolicy.DimensionGoFileLines &&
			indicator.MetaOperation == sourcepolicy.OperationSplitGo {
			indicators = append(indicators, indicator)
		}
	}
	sort.Slice(indicators, func(i, j int) bool { return indicators[i].Subject < indicators[j].Subject })
	return indicators
}

func containsSubject(indicators []metricIndicator, subject string) bool {
	for _, indicator := range indicators {
		if indicator.Subject == subject {
			return true
		}
	}
	return false
}
