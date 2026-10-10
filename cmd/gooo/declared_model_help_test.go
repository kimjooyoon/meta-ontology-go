package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDeclaredModelHelpShowsPracticalWorkflow(t *testing.T) {
	for _, topic := range []string{"models", "body-context"} {
		var out, diagnostics bytes.Buffer
		if code := run([]string{"help", topic}, &out, &diagnostics); code != exitOK || diagnostics.Len() != 0 {
			t.Fatal(topic, code, diagnostics.String())
		}
		for _, want := range []string{"model-choice-v1.json", "source-condition-cases.gooo.fixture", "body-context", "zero predictions"} {
			if !strings.Contains(out.String(), want) {
				t.Fatalf("%s help misses %q", topic, want)
			}
		}
		if topic == "models" {
			for _, want := range []string{
				"--path-model", "--model", "body-construct", "construction.json",
				"v0.2.37-experimental", "checksum", "deterministic", "docs/interaction-contract-model.md",
				"version --build --json", "model-interaction-v1.json", "v0.2.39-experimental",
				"source-interaction-condition-cases.gooo.fixture",
			} {
				if !strings.Contains(out.String(), want) {
					t.Fatalf("model workflow misses %q", want)
				}
			}
		}
	}
}
