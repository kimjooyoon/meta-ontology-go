package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestPackageGenerateImportedLibraries(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "go.mod", "module libraryprobe\n\ngo 1.27.3\n")
	fixture := t.TempDir()
	writeWorkspaceSource(t, fixture, "main.gooo", libraryHygieneSource)
	manifest := writeWorkspaceManifest(t, fixture, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"hygiene","activity":"Keep"},"packages":[{"path":"hygiene","sources":["main.gooo"]}]}`)
	if err := os.Mkdir(filepath.Join(root, "hygiene"), 0700); err != nil {
		t.Fatal(err)
	}
	writeWorkspaceSource(t, root, "hygiene/generated.go", string(generateLibraryCommand(t, "hygiene", manifest)))
	for _, example := range []struct{ name, folder, manifest string }{
		{"surface", "runtime-package-surface", "gooo.workspace.json"},
		{"optional", "package-optional-record-flow", "gooo.workspace.json"},
		{"named", "named-inputs", "workspace.json"},
	} {
		path := filepath.Join("..", "..", "examples", example.folder)
		generated := generateLibraryCommand(t, example.name, filepath.Join(path, example.manifest))
		if !bytes.Equal(generated, generateLibraryCommand(t, example.name, filepath.Join(path, example.manifest))) {
			t.Fatal("identical source did not reproduce identical library bytes")
		}
		if err := os.Mkdir(filepath.Join(root, example.name), 0700); err != nil {
			t.Fatal(err)
		}
		writeWorkspaceSource(t, root, example.name+"/generated.go", string(generated))
		if example.name != "named" {
			raw, err := os.ReadFile(filepath.Join(path, "cases.json"))
			if err != nil {
				t.Fatal(err)
			}
			writeWorkspaceSource(t, root, example.name+".json", string(raw))
		}
	}
	writeWorkspaceSource(t, root, "library_test.go", libraryConsumerOracle)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-count=1", "./...")
	command.Dir = root
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatal("fresh independent Go consumer:", err, string(out))
	}
}

func generateLibraryCommand(t *testing.T, name, manifest string) []byte {
	t.Helper()
	var out, diagnostics bytes.Buffer
	// These nested probes prove the generator does not depend on an available Go
	// executable. The separate native consumer test above performs the compilation.
	t.Run("generate-"+name, func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		if code := run([]string{"package", "generate", "--go-package", name, manifest}, &out, &diagnostics); code != exitOK || diagnostics.Len() != 0 {
			t.Fatal(code, diagnostics.String())
		}
	})
	for _, marker := range []string{"// Compiler version=", "// Workspace manifest: sha256:", "// Workspace image: sha256:",
		"// Gooo source package=", "// Lowered Gooo source: sha256:", "// Typed plan: sha256:", "func Evaluate(", "// Evaluate returns activity="} {
		if !strings.Contains(out.String(), marker) {
			t.Fatal("generated library lacks source binding", marker)
		}
	}
	return out.Bytes()
}

func TestPackageGenerateFailureHasNoSource(t *testing.T) {
	manifest := filepath.Join("..", "..", "examples", "runtime-package-surface", "gooo.workspace.json")
	for _, args := range [][]string{
		{"package", "generate"},
		{"package", "generate", "--go-package", "main", manifest},
		{"package", "generate", "--go-package", "bad-name", manifest},
		{"package", "generate", "--go-package", "rules", "missing.workspace.json"},
		{"package", "generate", "--go-package", "rules", manifest, "--cases", "ignored.json"},
	} {
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code == exitOK || out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatal("rejected generation emitted usable partial source", args, code, out.String(), diagnostics.String())
		}
	}
}

func TestPackageGenerateSourceDigestRepresentation(t *testing.T) {
	root := t.TempDir()
	source := libraryHygieneSource + "// 한국어 <>& \"quoted\"\n"
	writeWorkspaceSource(t, root, "main.gooo", source)
	manifest := writeWorkspaceManifest(t, root, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"hygiene","activity":"Keep"},"packages":[{"path":"hygiene","sources":["main.gooo"]}]}`)
	generated := generateLibraryCommand(t, "hygiene", manifest)
	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("sha256:%x", sha256.Sum256(encoded))
	raw := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(source)))
	if want == raw {
		t.Fatal("probe does not distinguish representations")
	}
	for _, line := range strings.Split(string(generated), "\n") {
		if strings.HasPrefix(line, "// Gooo source package=") {
			if !strings.Contains(line, fmt.Sprintf("source_json_string_sha256=%q", want)) || strings.Contains(line, raw) {
				t.Fatal("source hash must identify its JSON-string representation", line)
			}
			return
		}
	}
	t.Fatal("original source origin unavailable")
}
