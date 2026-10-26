package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageGenerateDirectoryPreservesPriorArtifactOnFailure(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, generatedFileName)
	original := []byte("package rules\n// earlier generated library\n")
	if err := os.WriteFile(path, original, 0640); err != nil {
		t.Fatal(err)
	}
	prior, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join("..", "..", "examples", "runtime-package-surface", "gooo.workspace.json")
	for _, input := range []struct{ name, path string }{{"bad-name", manifest}, {"rules", "missing.workspace.json"}} {
		var out, diagnostics bytes.Buffer
		code := run([]string{"package", "generate", "--go-package", input.name, "--out", root, input.path}, &out, &diagnostics)
		after, err := os.ReadFile(path)
		if code == exitOK || out.Len() != 0 || diagnostics.Len() == 0 || err != nil || !bytes.Equal(original, after) {
			t.Fatal("failed generation changed the prior artifact", code, err, diagnostics.String())
		}
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "generate", manifest, "--go-package", "rules", "--out", root}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	if out.String() != fmt.Sprintf("generated: %s\n", path) || diagnostics.Len() != 0 {
		t.Fatal("publication result was not delivered", out.String(), diagnostics.String())
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != prior.Mode().Perm() {
		t.Fatal("replacement lost the prior mode", info, err)
	}
	if code := run([]string{"package", "generate", "--go-package", "rules", "--out", root, manifest}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	again, err := os.Stat(path)
	if err != nil || !os.SameFile(info, again) || !info.ModTime().Equal(again.ModTime()) {
		t.Fatal("unchanged generation replaced the file", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || entries[0].Name() != generatedFileName {
		t.Fatal("publication leaked temporary files", entries, err)
	}
}

func TestPackageGenerateDirectoryRejectsWorkspaceInputTargets(t *testing.T) {
	for _, member := range []string{"source", "manifest"} {
		t.Run(member, func(t *testing.T) {
			root := t.TempDir()
			sourceName, manifestName := "main.gooo", "workspace.json"
			if member == "manifest" {
				manifestName = generatedFileName
			}
			definition := fmt.Sprintf(`{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"hygiene","activity":"Keep"},"packages":[{"path":"hygiene","sources":[%q]}]}`, sourceName)
			originals := map[string]string{sourceName: libraryHygieneSource, manifestName: definition}
			for name, source := range originals {
				writeWorkspaceSource(t, root, name, source)
			}
			if member == "source" {
				if err := os.Link(filepath.Join(root, sourceName), filepath.Join(root, generatedFileName)); err != nil {
					t.Fatal(err)
				}
				originals[generatedFileName] = libraryHygieneSource
			}
			var out, diagnostics bytes.Buffer
			if code := run([]string{"package", "generate", "--go-package", "rules", "--out", root,
				filepath.Join(root, manifestName)}, &out, &diagnostics); code == exitOK || out.Len() != 0 ||
				!strings.Contains(diagnostics.String(), "would replace workspace input") {
				t.Fatal(code, out.String(), diagnostics.String())
			}
			for name, original := range originals {
				after, err := os.ReadFile(filepath.Join(root, name))
				if err != nil || string(after) != original {
					t.Fatal("generation changed its input", name, err)
				}
			}
		})
	}
}

func TestPackageGenerateDirectoryReportFailureRetainsPublishedFile(t *testing.T) {
	manifest := filepath.Join("..", "..", "examples", "runtime-package-surface", "gooo.workspace.json")
	want := generateLibraryCommand(t, "rules", manifest)
	root := filepath.Join(t.TempDir(), "new", "rules")
	writer := &generateDeliveryWriter{failAt: 1}
	var diagnostics bytes.Buffer
	if code := run([]string{"package", "generate", "--go-package", "rules", "--out", root, manifest}, writer, &diagnostics); code != exitFailure {
		t.Fatal("report delivery failure was ignored", code)
	}
	got, err := os.ReadFile(filepath.Join(root, generatedFileName))
	if err != nil || !bytes.Equal(got, want) || writer.calls != 1 {
		t.Fatal("failed success message changed the completed file", err, writer.calls)
	}
}

func TestPackageGenerateDirectoryFailuresCreateNoOutput(t *testing.T) {
	manifest := filepath.Join("..", "..", "examples", "runtime-package-surface", "gooo.workspace.json")
	root := t.TempDir()
	directory := filepath.Join(root, "not-created", "rules")
	for _, extra := range [][]string{{"--out"}, {"--out", ""}, {"--out", directory, "--out", directory},
		{"--out", directory, "--go-package", "other"}, {"--unknown"}} {
		args := append([]string{"package", "generate", "--go-package", "rules", manifest}, extra...)
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitUsage || out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatal(code, out.String(), diagnostics.String())
		}
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "generate", "--go-package", "bad-name", "--out", directory, manifest}, &out, &diagnostics); code == exitOK {
		t.Fatal("invalid generation succeeded")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatal("failed generation created output directories", entries, err)
	}
	if err := os.Mkdir(filepath.Join(root, generatedFileName), 0755); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := run([]string{"package", "generate", "--go-package", "rules", "--out", root, manifest}, &out, &diagnostics); code == exitOK || out.Len() != 0 {
		t.Fatal("non-file target was published", code, out.String())
	}
	entries, err = os.ReadDir(root)
	if err != nil || len(entries) != 1 || !entries[0].IsDir() {
		t.Fatal("failed publication changed its directory", entries, err)
	}
}
