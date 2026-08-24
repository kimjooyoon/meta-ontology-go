package languageassurance

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityactivation"
)

func operatingOperationSet() (map[string]string, error) {
	operations := make(map[string]string, len(baseOperatingOperations)+1)
	for metricID, operation := range baseOperatingOperations {
		operations[metricID] = operation
	}
	metricID, operation, err := sourceauthorityactivation.OperatingOperation()
	if err != nil {
		return nil, fmt.Errorf("source authority activation: %w", err)
	}
	if metricID != MetricSourceAuthority || operation != "bind-source-backed-authority" {
		return nil, fmt.Errorf("source authority activation operation mismatch")
	}
	operations[metricID] = operation
	return operations, nil
}
