package selectiveci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

func normalizeResult(result PlanResult) PlanResult {
	result.ChangedSemanticIDs = sortedUnique(result.ChangedSemanticIDs)
	result.SelectedCommandIDs = sortedUnique(result.SelectedCommandIDs)
	result.SelectedGuardCommandIDs = sortedUnique(result.SelectedGuardCommandIDs)
	result.SelectedWorkIDs = sortedUnique(result.SelectedWorkIDs)
	result.ResourceReceiptDigests = sortedUnique(result.ResourceReceiptDigests)
	result.ProvenancePathIDs = sortedUnique(result.ProvenancePathIDs)
	return result
}
func sortedCopy(values []string) []string {
	if values == nil {
		return nil
	}
	copy := append([]string{}, values...)
	sort.Strings(copy)
	return copy
}
func sortedUnique(values []string) []string {
	values = sortedCopy(values)
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
func edgeKey(edge DependencyEdge) string {
	return edge.From + "\x00" + string(edge.Kind) + "\x00" + edge.To
}
func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}
