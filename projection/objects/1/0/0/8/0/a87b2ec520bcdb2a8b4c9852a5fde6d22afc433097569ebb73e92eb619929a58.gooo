package metricevidence

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type Report struct {
	CommitSHA   string      `json:"commit_sha"`
	Directories []Directory `json:"directories"`
	Meta        struct {
		Policy struct {
			MaxFileLines     int `json:"max_file_lines"`
			MaxDirectEntries int `json:"max_direct_directory_entries"`
		} `json:"policy"`
		Indicators []Indicator `json:"indicators"`
	} `json:"meta"`
}

type Directory struct {
	Path          string `json:"path"`
	DirectFolders int    `json:"direct_folders"`
	DirectFiles   int    `json:"direct_files"`
}

type Indicator struct {
	MetricID      sourcepolicy.Dimension `json:"metric_id"`
	Subject       string                 `json:"subject"`
	Satisfied     bool                   `json:"satisfied"`
	MetaOperation sourcepolicy.Operation `json:"meta_operation"`
}

func Load(path, expectedSHA string) (Report, error) {
	var report Report
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
