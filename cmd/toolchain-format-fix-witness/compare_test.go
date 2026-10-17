package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
)

func TestSavedReportComparisonIsReadOnly(t *testing.T) {
	const head = "0000000000000000000000000000000000000000"
	report := metaff.Evaluate(metaff.Input{ExpectedHeadSHA: head})
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	first, replay := filepath.Join(root, "first.json"), filepath.Join(root, "replay.json")
	for _, path := range []string{first, replay} {
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := run(config{compare: first, check: replay, head: head}); err != nil {
		t.Fatal(err)
	}
	if err := writeOrCheck("", first, report); err != nil {
		t.Fatal("fresh report comparison rejected equivalent JSON formatting", err)
	}
	for _, path := range []string{first, replay} {
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(raw, after) {
			t.Fatal("saved comparison changed a raw report", path, err)
		}
	}
	if report.Decision != metaff.DecisionClosed {
		t.Fatal("fixture must demonstrate equality without successful execution authority")
	}
}

func TestSavedReportComparisonRejectsExecutionConfiguration(t *testing.T) {
	changes := []struct {
		name   string
		mutate func(*config)
	}{
		{"check", func(c *config) { c.check = "" }},
		{"head", func(c *config) { c.head = "" }},
		{"root", func(c *config) { c.root = "root" }},
		{"executable", func(c *config) { c.executable = "executable" }},
		{"concept", func(c *config) { c.concept = "concept" }},
		{"corpus", func(c *config) { c.corpus = "corpus" }},
		{"output", func(c *config) { c.output = "output" }},
	}
	for _, change := range changes {
		t.Run(change.name, func(t *testing.T) {
			cfg := config{compare: "unopened-first.json", check: "unopened-replay.json", head: "head"}
			change.mutate(&cfg)
			want := "saved comparison accepts no execution or output configuration"
			if change.name == "check" || change.name == "head" {
				want = "compare, check, and expected-head are required for saved report comparison"
			}
			if err := run(cfg); err == nil || err.Error() != want {
				t.Fatal("configuration was not rejected before report reads", change.name, err)
			}
		})
	}
}
