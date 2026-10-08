package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func metadataWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeWorkspaceSource(t, root, "app.gooo", `package app
namespace app
import nums "std/numbers"
entity Integer id "metadata://integer"
activity Run(Integer) -> Integer computes "return input"
`)
	writeWorkspaceSource(t, root, "more.gooo", `package app
namespace app
import "std/text"
import "std/numbers"
entity Text id "metadata://text"
`)
	writeWorkspaceSource(t, root, "numbers.gooo", `package numbers
namespace numbers
entity Integer id "metadata://integer"
`)
	writeWorkspaceSource(t, root, "text.gooo", `package text
namespace text
entity Text id "metadata://text"
`)
	return writeWorkspaceManifest(t, root, `{
  "schema":"gooo/package-workspace-manifest/v1",
  "entry":{"package_path":"example/app","activity":"Run"},
  "packages":[
    {"path":"example/app","sources":["more.gooo","app.gooo"]},
    {"path":"std/text","sources":["text.gooo"]},
    {"path":"std/numbers","sources":["numbers.gooo"]}
  ]
}`)
}

func resolveMetadata(t *testing.T, path string, want int) workspaceResolutionReceipt {
	t.Helper()
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "resolve", "--json", path}, &out, &diagnostics)
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != want || diagnostics.Len() != 0 {
		t.Fatal(code, err, out.String(), diagnostics.String())
	}
	return receipt
}

func TestWorkspaceSourceMetadataMatchesExplicitGraph(t *testing.T) {
	path := metadataWorkspace(t)
	derived := resolveMetadata(t, path, exitOK)
	if derived.Result == nil || derived.Decision != "PASS" {
		t.Fatal("missing source-derived package graph")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var explicit workspaceManifest
	if err = json.Unmarshal(raw, &explicit); err != nil {
		t.Fatal(err)
	}
	for i := range explicit.Packages {
		p := &explicit.Packages[i]
		p.Name = filepath.Base(p.Path)
		p.Imports = []string{}
		if p.Name == "app" {
			p.Imports = []string{"std/numbers", "std/text"}
		}
	}
	raw, err = json.Marshal(explicit)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	declared := resolveMetadata(t, path, exitOK)
	if !reflect.DeepEqual(derived.Result, declared.Result) || derived.ManifestDigest == declared.ManifestDigest {
		t.Fatal("equivalent source metadata changed the graph or lost manifest identity")
	}
}

func TestWorkspaceSourceMetadataRejectsInconsistentSources(t *testing.T) {
	for _, tc := range []struct{ name, source, code string }{
		{"package", "package other\nnamespace other\nentity Text id \"metadata://text\"\n", "PACKAGE_HEADER_MISMATCH"},
		{"syntax", "package app\nnamespace app\nactivity !!!", "PACKAGE_SOURCE_INVALID"},
		{"dependency", "package app\nnamespace app\nimport \"missing/package\"\n", "PACKAGE_IMPORT_UNKNOWN"},
		{"cycle", "package app\nnamespace app\nimport \"example/app\"\n", "PACKAGE_IMPORT_CYCLE"},
		{"header", "package app\nentity Text id \"metadata://text\"\n", "PACKAGE_SOURCE_INVALID"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := metadataWorkspace(t)
			writeWorkspaceSource(t, filepath.Dir(path), "more.gooo", tc.source)
			r := resolveMetadata(t, path, exitFailure)
			if r.Decision != "FAIL_CLOSED" || !strings.Contains(r.Error, tc.code) {
				t.Fatal(r.Error)
			}
		})
	}
}
