package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupGoSourceAvailabilityBinding(t *testing.T) {
	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "docs/observations/action-origin-upgrade/observations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var origin struct {
		Revision string `json:"selected_revision"`
		Release  string `json:"selected_release"`
		Runtime  string `json:"node_runtime"`
	}
	if err := json.Unmarshal(data, &origin); err != nil || origin.Runtime != "node24" || len(origin.Revision) != 40 {
		t.Fatal("missing selected action origin", err)
	}
	paths, err := filepath.Glob(filepath.Join(root, ".github/workflows/*.yml"))
	if err != nil {
		t.Fatal(err)
	}
	sites := 0
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, site := range parseWorkflow(source) {
			if site.Action == "actions/setup-go" {
				sites++
				if site.Ref != origin.Revision && site.Ref != origin.Release {
					t.Fatalf("%s:%d: setup-go origin %s is outside the selected revision/release", path, site.Line, site.Ref)
				}
			}
		}
	}
	if sites == 0 {
		t.Fatal("no setup-go consumers were checked")
	}
}
