package languageassurance

import (
	"fmt"
	"maps"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityactivation"
)

func operatingOperationSet() (map[string]string, error) {
	operations := make(map[string]string, len(operatingOperations)+1)
	maps.Copy(operations, operatingOperations)
	metricID, operation, err := sourceauthorityactivation.OperatingOperation()
	if err != nil {
		return nil, fmt.Errorf("source authority activation: %w", err)
	}
	if metricID != "gooo.metric.semantic.source-backed-authority.v1" || operation != "bind-source-backed-authority" {
		return nil, fmt.Errorf("source authority activation operation mismatch")
	}
	operations[metricID] = operation
	return operations, nil
}

func SnapshotEvidenceIDs() []string { return append([]string(nil), snapshotEvidenceIDs...) }
