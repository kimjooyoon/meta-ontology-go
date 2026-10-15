package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadableOriginsCLI133613(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plain133613.gooo")
	source := "package plain\nnamespace plain\nentity Integer id \"urn:gooo:type:integer\"\nactivity Sum(Integer) -> Integer id \"urn:plain133613:sum\" computes `let subtotal=input+239;return subtotal`\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var text, diagnostics bytes.Buffer
	if status := run([]string{"body-origins", "--activity", "Sum", "--format", "text", path}, &text, &diagnostics); status != exitOK || !strings.Contains(text.String(), "subtotal") || !strings.Contains(text.String(), "Scope: possible source dependencies") {
		t.Fatal(status, text.String(), diagnostics.String())
	}
	var machine bytes.Buffer
	if status := run([]string{"body-origins", "--activity", "Sum", path}, &machine, &diagnostics); status != exitOK || !json.Valid(machine.Bytes()) {
		t.Fatal(status, machine.String(), diagnostics.String())
	}
	var rejected bytes.Buffer
	if status := run([]string{"body-origins", "--activity", "Sum", "--format", "text", "--inputs-for", "plain:Inputs", path}, &rejected, &diagnostics); status != exitUsage || rejected.Len() != 0 {
		t.Fatal(status, rejected.String(), diagnostics.String())
	}
}
