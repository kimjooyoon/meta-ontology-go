package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOriginStarterGuideShowsValuesAndReceipts(t *testing.T) {
	for _, module := range []string{"origins", "example.org/team/reader"} {
		destination := filepath.Join(t.TempDir(), "reader")
		args := []string{"init", "--template", "origins"}
		if module != "origins" {
			args = append(args, "--module", module)
		}
		var out, diagnostics bytes.Buffer
		if code := run(append(args, destination), &out, &diagnostics); code != exitOK {
			t.Fatal("origin project creation", code, diagnostics.String())
		}
		guide, err := os.ReadFile(filepath.Join(destination, "README.md"))
		if err != nil {
			t.Fatal(err)
		}
		for _, text := range []string{module, "gooo package execute --cases cases.json", "gooo package execute --inputs inputs.json", "one JSON entry value per input row", "With `--json`"} {
			if !bytes.Contains(guide, []byte(text)) {
				t.Fatal("generated guide lost the value/record distinction", text)
			}
		}
		manifest, err := os.ReadFile(filepath.Join(destination, "gooo.workspace.json"))
		if err != nil {
			t.Fatal(err)
		}
		var workspace struct {
			Entry struct {
				PackagePath string `json:"package_path"`
				Activity    string `json:"activity"`
			} `json:"entry"`
		}
		if err := json.Unmarshal(manifest, &workspace); err != nil || workspace.Entry.PackagePath != module+"/origins" || workspace.Entry.Activity != "Inspect" {
			t.Fatal("starter entry identity", workspace, err)
		}
	}
	var help, diagnostics bytes.Buffer
	if code := run([]string{"help", "package"}, &help, &diagnostics); code != exitOK ||
		!strings.Contains(help.String(), "Successful input-only execution prints one JSON value per input row.") {
		t.Fatal("package help lost the direct-value entry point", code, diagnostics.String())
	}
	t.Log("default and explicit origin modules create guides and stable entries; no Gooo program or model executed")
}
