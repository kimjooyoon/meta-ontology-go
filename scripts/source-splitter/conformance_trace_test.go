package main

import (
	"path/filepath"
	"testing"

	conformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
)

func normalizeSplitEvents(t *testing.T, root string, observed []splitEvent) []conformance.WriteEvent {
	t.Helper()
	result := make([]conformance.WriteEvent, len(observed))
	for index, event := range observed {
		target := relativeTracePath(t, root, event.Target)
		temporary := ""
		if event.Temporary != "" {
			temporary = relativeTracePath(t, root, event.Temporary)
		}
		result[index] = conformance.WriteEvent{Sequence: index, Kind: event.Kind,
			Target: target, Temporary: temporary, Success: event.Success}
	}
	return result
}

func relativeTracePath(t *testing.T, root, value string) string {
	t.Helper()
	if value == "" {
		return ""
	}
	relative, err := filepath.Rel(root, value)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(relative)
}
