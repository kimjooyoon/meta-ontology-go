package pathclosure

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func indexR4Records(values []R4Record) (map[semantic.ID]R4Record, error) {
	result := make(map[semantic.ID]R4Record, len(values))
	for _, raw := range values {
		record, err := normalizeR4Record(raw)
		if err != nil {
			return nil, err
		}
		if _, exists := result[record.ID]; exists {
			return nil, fmt.Errorf("duplicate record %s", record.ID)
		}
		result[record.ID] = record
	}
	return result, nil
}
func indexR4Receipts(values []R4Receipt) (map[semantic.ID]R4Receipt, error) {
	result := make(map[semantic.ID]R4Receipt, len(values))
	events := make(map[semantic.ID]struct{}, len(values))
	for _, raw := range values {
		receipt, err := normalizeR4Receipt(raw)
		if err != nil {
			return nil, err
		}
		if receipt.ID == "" || receipt.EventID == "" {
			return nil, fmt.Errorf("receipt and append-only event IDs are required")
		}
		if _, exists := result[receipt.ID]; exists {
			return nil, fmt.Errorf("duplicate receipt %s", receipt.ID)
		}
		if _, exists := events[receipt.EventID]; exists {
			return nil, fmt.Errorf("conflicting append-only event %s", receipt.EventID)
		}
		result[receipt.ID] = receipt
		events[receipt.EventID] = struct{}{}
	}
	return result, nil
}
func indexR4Paths(values []R4Path) (map[semantic.ID]R4Path, error) {
	result := make(map[semantic.ID]R4Path, len(values))
	for _, raw := range values {
		path, err := normalizeR4Path(raw)
		if err != nil {
			return nil, err
		}
		if _, exists := result[path.ID]; exists {
			return nil, fmt.Errorf("duplicate path %s", path.ID)
		}
		result[path.ID] = path
	}
	return result, nil
}
