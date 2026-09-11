package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSourceBindsAuthoritativeWorkflowDiscoveryPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source-metrics.json")
	data := []byte(`{"meta":{"policy":{"exempt_workflow_discovery_root":true}}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := loadSource(path)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Meta.Policy.ExemptWorkflowDiscoveryRoot {
		t.Fatal("workflow discovery root policy was not bound")
	}
}
