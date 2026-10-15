package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentCommandManifestSourceOrigins(t *testing.T) {
	root := filepath.Join("..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "examples/ci-effort-observation/operations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil || len(manifest.Operations) != 7 {
		t.Fatal("current operation scope", err)
	}
	workflow, err := os.ReadFile(filepath.Join(root, manifest.WorkflowSource))
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range manifest.Operations {
		for _, event := range []string{"pull_request", "push"} {
			current, err := bindWorkflowCommandForEvent(workflow, manifest.WorkflowSource, spec, event)
			if err != nil || current == "" {
				t.Fatal("declared operation lacks exact current source", spec.ID, event, err)
			}
			if spec.ID != "go-test-race" {
				continue
			}
			legacy := spec
			legacy.StepName = "Run go test -race ./..."
			legacy.Command = []string{"go", "test", "-race", "./..."}
			if _, err := bindWorkflowCommandForEvent(workflow, manifest.WorkflowSource, legacy, event); err == nil {
				t.Fatal("legacy command inherited the new source context")
			}
			oldWorkflow := []byte(strings.ReplaceAll(string(workflow), "go test -race -json ./internal/meta/repositoryprojection/extractor ./...", "go test -race ./..."))
			previous, err := bindWorkflowCommandForEvent(oldWorkflow, manifest.WorkflowSource, legacy, event)
			if err != nil || previous == current {
				t.Fatal("command/source change lost context identity", err)
			}
		}
	}
}
