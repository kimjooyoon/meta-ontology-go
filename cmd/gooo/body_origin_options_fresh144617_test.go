package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOriginOptionsAndHelperText144617(t *testing.T) {
	options, err := parseBodyOriginOptions([]string{"--activity", "Observe144617", "--inputs-for", "audit144617:Read", "--include-control", "source144617.gooo"})
	if err != nil || !options.control || options.format != "json" || options.file != "source144617.gooo" {
		t.Fatal(options, err)
	}
	if _, err = parseBodyOriginOptions([]string{"--activity", "Observe144617", "--format", "json", "--format", "text", "source144617.gooo"}); err == nil {
		t.Fatal("repeated output format accepted")
	}
	path := filepath.Join(t.TempDir(), "helper144617.gooo")
	source := "package view144617\nnamespace view144617\nentity Text id \"urn:gooo:type:text\"\nentity Boolean id \"urn:gooo:type:boolean\"\nactivity Decorate(Text) -> Text id \"urn:view144617:decorate\" computes `return \"[\"+input+\"]\"`\nactivity Show(Text,Boolean) -> Text id \"urn:view144617:show\" computes `let title=Decorate(input0);if input1{return title+\"urgent144617\"};return title`\n"
	if err = os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var output, diagnostics bytes.Buffer
	status := run([]string{"body-origins", "--activity", "Show", "--format", "text", path}, &output, &diagnostics)
	if status != exitOK || !strings.Contains(output.String(), "Decorate") || !strings.Contains(output.String(), "control <- #") || !strings.Contains(output.String(), "urn:view144617:show") {
		t.Fatal(status, output.String(), diagnostics.String())
	}
}
