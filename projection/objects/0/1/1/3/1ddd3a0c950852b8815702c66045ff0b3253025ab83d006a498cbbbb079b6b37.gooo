package workfrontier

import (
	"encoding/json"
	"testing"
)

type observedResult struct {
	Status      string
	Quality     string
	SelectedIDs []string
	WorkIDs     []string
}

func observeResult(t *testing.T, result any) observedResult {
	t.Helper()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	return observedResult{
		Status:      exactString(t, object, "status"),
		Quality:     exactStringOptional(t, object, "quality"),
		SelectedIDs: exactStrings(t, object, "selected_ids"),
		WorkIDs:     exactStrings(t, object, "work_ids"),
	}
}
func exactString(t *testing.T, object map[string]any, key string) string {
	t.Helper()
	value, ok := object[key].(string)
	if !ok {
		t.Fatalf("result field %q is missing or not a string: %#v", key, object[key])
	}
	return value
}
func exactStringOptional(t *testing.T, object map[string]any, key string) string {
	t.Helper()
	if _, ok := object[key]; !ok {
		return ""
	}
	return exactString(t, object, key)
}
func exactStrings(t *testing.T, object map[string]any, key string) []string {
	t.Helper()
	value, ok := object[key]
	if !ok || value == nil {
		return []string{}
	}
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("result field %q is not an array: %#v", key, value)
	}
	result := make([]string, len(items))
	for i, item := range items {
		result[i], ok = item.(string)
		if !ok {
			t.Fatalf("result field %q item %d is not a string: %#v", key, i, item)
		}
	}
	return result
}
