package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStarterAgentGuideIsLocalAndDoesNotMutateTemplates(t *testing.T) {
	for _, template := range []string{"app", "library", "diagnostic"} {
		t.Run(template, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), template)
			var out, diagnostics bytes.Buffer
			args := []string{"init", "--template", template, root}
			if template != "app" {
				args = append(args, "--module", "example.org/team/tool")
			}
			if code := run(args, &out, &diagnostics); code != exitOK {
				t.Fatalf("init %s: %d %s", template, code, &diagnostics)
			}
			guide, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
			if err != nil || string(guide) != starterAgentGuide || strings.Contains(string(guide), "{{") {
				t.Fatalf("local guide differs from embedded guidance: %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, "README.md")); err != nil {
				t.Fatal("guide's local README reference is missing", err)
			}
			if template == "app" {
				out.Reset()
				args = []string{"context", filepath.Join(root, "main.gooo"), "starter://activity/clamp"}
			} else {
				out.Reset()
				args = []string{"package", "interface", "--json", filepath.Join(root, "gooo.workspace.json")}
			}
			if code := run(args, &out, &diagnostics); code != exitOK || out.Len() == 0 {
				t.Fatalf("guide's inspection route is unavailable: %d %s", code, &diagnostics)
			}
		})
	}
	for _, files := range []map[string]string{starterFiles, libraryStarterFiles, diagnosticStarterFiles} {
		if _, found := files["AGENTS.md"]; found {
			t.Fatal("selection mutated the shared source template")
		}
	}
}
