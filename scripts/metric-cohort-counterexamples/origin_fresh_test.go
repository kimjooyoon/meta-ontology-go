package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOriginalProducerOriginFresh1677(t *testing.T) {
	for _, row := range []struct {
		name   string
		mutate func(*config, *artifactMetadata)
		want   string
	}{
		{"exact", func(*config, *artifactMetadata) {}, ""},
		{"wrong-id", func(_ *config, m *artifactMetadata) { m.ID++ }, "metadata differs"},
		{"wrong-run", func(_ *config, m *artifactMetadata) { m.WorkflowRun.ID++ }, "metadata differs"},
		{"wrong-source", func(_ *config, m *artifactMetadata) { m.WorkflowRun.HeadSHA = strings.Repeat("b", 40) }, "metadata differs"},
		{"wrong-name", func(_ *config, m *artifactMetadata) { m.Name += "-other" }, "metadata differs"},
		{"expired", func(_ *config, m *artifactMetadata) { m.Expired = true }, "metadata differs"},
		{"rerun", func(c *config, _ *artifactMetadata) { c.Origin.Attempt = 2 }, "exact original"},
	} {
		t.Run(row.name, func(t *testing.T) {
			cfg := transportFixtureFresh1677(t, transportNames())
			raw, err := os.ReadFile(cfg.Metadata)
			if err != nil {
				t.Fatal(err)
			}
			var metadata artifactMetadata
			if err := json.Unmarshal(raw, &metadata); err != nil {
				t.Fatal(err)
			}
			row.mutate(&cfg, &metadata)
			raw, err = json.Marshal(metadata)
			if err != nil {
				t.Fatal(err)
			}
			writeFixtureFresh1677(t, cfg.Metadata, raw)
			_, err = boundArchive(cfg)
			if row.want == "" && err != nil || row.want != "" && (err == nil || !strings.Contains(err.Error(), row.want)) {
				t.Fatalf("original producer identity: %v", err)
			}
		})
	}
}
