package extractor

import (
	"bytes"
	"context"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCallbackObservationPreservesModuleAndPackageIdentityCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("native module ownership regression is CI-only")
	}
	root, output := t.TempDir(), t.TempDir()
	const logical = "cmd/subject/subject_test.go"
	files := map[string][]byte{
		"go.mod":                         []byte("module example.invalid/callback-observation/v2\n\ngo 1.27.0\n"),
		"go.sum":                         []byte(""),
		"internal/value/value.go":        []byte("package value\nconst Answer = 42\n"),
		"cmd/subject/testdata/proof.txt": []byte("original-input"),
		logical: []byte(`package subject
import (
	"os"
	"reflect"
	"testing"
	"example.invalid/callback-observation/v2/internal/value"
)
type identity struct{}
func TestRequired(t *testing.T) {
	if reflect.TypeOf(identity{}).PkgPath() != "example.invalid/callback-observation/v2/cmd/subject" || value.Answer != 42 {
		t.Fatal("original package identity or internal import changed")
	}
	data, err := os.ReadFile("testdata/proof.txt")
	if err != nil || string(data) != "original-input" {
		t.Fatalf("package data changed: %s %v", data, err)
	}
}
`),
	}
	baseline := map[string][]byte{}
	for name, raw := range files {
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		if relative, ok := strings.CutPrefix(name, "cmd/subject/"); ok {
			baseline[relative] = raw
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("excluded worktree metadata"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	snapshot, err := callbackObservationModuleSources(ctx, root, logical, baseline)
	if err != nil || !maps.EqualFunc(snapshot, files, bytes.Equal) {
		t.Fatalf("module snapshot changed input bytes: %v", err)
	}
	for _, variant := range []string{"source", "final"} {
		packageFiles := maps.Clone(baseline)
		if variant == "final" {
			packageFiles["generated.go"] = []byte("package subject\nconst Generated = true\n")
		}
		directory, err := materializeCallbackObservation(output, variant, logical, snapshot, packageFiles)
		if err != nil {
			t.Fatal(err)
		}
		goMod, err := os.ReadFile(filepath.Join(output, variant, "go.mod"))
		if err != nil || !bytes.Equal(goMod, files["go.mod"]) {
			t.Fatalf("original module declaration was rewritten: %v", err)
		}
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); !os.IsNotExist(err) {
			t.Fatalf("overlapping package module was created: %v", err)
		}
		run, err := runCallbackPackageObservation(ctx, directory, variant, "TestRequired")
		if err != nil || run.ExitCode != 0 || !run.TestEventsComplete {
			t.Fatalf("original module/package identity failed: %v; stdout=%s; stderr=%s", err, run.Stdout, run.Stderr)
		}
		t.Logf("module identity regression variant=%s exit=%d wall_ms=%d events=%d", variant, run.ExitCode, run.WallMS, len(run.Events))
	}
	for name, expected := range files {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil || !bytes.Equal(raw, expected) {
			t.Fatalf("input repository was modified at %s: %v", name, err)
		}
	}
}

func TestCallbackObservationModuleSnapshotRejectsDriftAndCancellation(t *testing.T) {
	root := t.TempDir()
	for name, raw := range map[string]string{"go.mod": "module example.invalid/snapshot\n\ngo 1.27.0\n", "subject_test.go": "package subject\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := callbackObservationModuleSources(context.Background(), root, "subject_test.go", map[string][]byte{}); err == nil {
		t.Fatal("package drift was accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := callbackObservationModuleSources(ctx, root, "subject_test.go", nil); err != context.Canceled {
		t.Fatalf("module snapshot ignored its context: %v", err)
	}
}
