package languageassurance

import (
	"fmt"
	"maps"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/candidateleakageactivation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/changedsurfacereceiptactivation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/rollbackintegrityactivation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityactivation"
)

func operatingOperationSet() (map[string]string, error) {
	operations := make(map[string]string, len(operatingOperations)+4)
	maps.Copy(operations, operatingOperations)
	metricID, operation, err := sourceauthorityactivation.OperatingOperation()
	if err != nil {
		return nil, fmt.Errorf("source authority activation: %w", err)
	}
	if metricID != "gooo.metric.semantic.source-backed-authority.v1" || operation != "bind-source-backed-authority" {
		return nil, fmt.Errorf("source authority activation operation mismatch")
	}
	operations[metricID] = operation
	metricID, operation, err = candidateleakageactivation.OperatingOperation()
	if err != nil {
		return nil, fmt.Errorf("candidate leakage activation: %w", err)
	}
	if metricID != "gooo.metric.semantic.candidate-leakage.v1" || operation != "detect-candidate-leakage" {
		return nil, fmt.Errorf("candidate leakage activation operation mismatch")
	}
	operations[metricID] = operation
	metricID, operation, err = changedsurfacereceiptactivation.OperatingOperation()
	if err != nil {
		return nil, fmt.Errorf("changed surface receipt activation: %w", err)
	}
	if metricID != "gooo.metric.semantic.changed-surface-receipt-totality.v1" || operation != "totalize-changed-surface-receipts" {
		return nil, fmt.Errorf("changed surface receipt activation operation mismatch")
	}
	operations[metricID] = operation
	metricID, operation, err = rollbackintegrityactivation.OperatingOperation()
	if err != nil {
		return nil, fmt.Errorf("rollback integrity activation: %w", err)
	}
	if metricID != "gooo.metric.operation.rollback-integrity.v1" || operation != "verify-rollback-integrity" {
		return nil, fmt.Errorf("rollback integrity activation operation mismatch")
	}
	operations[metricID] = operation
	return operations, nil
}

func SnapshotEvidenceIDs() []string { return append([]string(nil), snapshotEvidenceIDs...) }
