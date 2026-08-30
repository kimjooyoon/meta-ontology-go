package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestTopologyFailuresExemptsWorkflowDiscoveryRoot(t *testing.T) {
	root := t.TempDir()
	workflowRoot := filepath.Join(root, ".github", "workflows")
	otherRoot := filepath.Join(root, "other")
	if err := os.MkdirAll(workflowRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(otherRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 11; index++ {
		if err := os.WriteFile(filepath.Join(workflowRoot, fmt.Sprintf("workflow-%02d.yml", index)), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(otherRoot, fmt.Sprintf("object-%02d.blob", index)), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	direct, mixed, err := topologyFailures(root)
	if err != nil {
		t.Fatal(err)
	}
	if direct != 1 || mixed != 0 {
		t.Fatalf("topology failures = direct %d, mixed %d; want direct 1, mixed 0", direct, mixed)
	}
}
