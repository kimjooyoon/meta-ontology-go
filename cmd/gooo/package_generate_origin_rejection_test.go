package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageGenerateEmbeddedOriginRejectsWithoutPublishing(t *testing.T) {
	fixture, destination := t.TempDir(), t.TempDir()
	writeWorkspaceSource(t, fixture, "main.gooo", originProbeSource)
	manifest := writeWorkspaceManifest(t, fixture, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origins","activity":"Divide"},"packages":[{"path":"origins","sources":["main.gooo"]}]}`)
	writeWorkspaceSource(t, destination, generatedFileName, "prior bytes")
	for _, flags := range [][]string{
		{"--embed-origin", "--embed-origin"},
		{"--embed-origin", "--export-stage", "GoooGenerationOrigin=urn:origin:divide"},
	} {
		var out, diagnostics bytes.Buffer
		args := append([]string{"package", "generate", "--go-package", "origins", "--out", destination, manifest}, flags...)
		if code := run(args, &out, &diagnostics); code == exitOK || out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatal(code, out.String(), diagnostics.String())
		}
		prior, err := os.ReadFile(filepath.Join(destination, generatedFileName))
		if err != nil || string(prior) != "prior bytes" {
			t.Fatal("origin rejection replaced prior source", err)
		}
	}
}
