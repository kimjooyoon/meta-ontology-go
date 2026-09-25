package directorypartition

import (
	"encoding/json"
	"fmt"
)

const (
	indicatorSchema = "gooo/indicator-report/v3"
	policySchema    = "gooo/source-policy/v1"
)

func DecodeSource(payload []byte) (SourceMetrics, error) {
	var source SourceMetrics
	if err := json.Unmarshal(payload, &source); err != nil {
		return SourceMetrics{}, err
	}
	if err := validateSource(source); err != nil {
		return SourceMetrics{}, err
	}
	return source, nil
}

func validateSource(source SourceMetrics) error {
	if source.Repository == "" || source.CommitSHA == "" {
		return fmt.Errorf("source metrics exact subject is incomplete")
	}
	if source.Meta.Schema != indicatorSchema || source.Meta.Policy.Schema != policySchema {
		return fmt.Errorf("source metrics schema is unsupported")
	}
	if source.Meta.Policy.MaxDirectDirectoryEntries <= 0 {
		return fmt.Errorf("directory entry limit must be positive")
	}
	for _, directory := range source.Directories {
		if directory.Path == "." && directory.SubjectKind == "PROJECT_ROOT" {
			return nil
		}
	}
	return fmt.Errorf("source metrics project root is missing")
}
