package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCommittedVersionProjectionMatchesGooo(t *testing.T) {
	if err := project("../..", false); err != nil {
		t.Fatal(err)
	}
}

func TestVersionProjectionPreservesPriorOutputUntilExplicitRegeneration(t *testing.T) {
	root, original := projectionFixture(t)
	if err := project(root, true); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, outputPath))
	if err != nil || project(root, false) != nil {
		t.Fatal("initial projection check", err)
	}
	literal := regexp.MustCompile(`"[0-9]+\.[0-9]+\.[0-9]+-dev"`).FindString(string(original))
	if literal == "" {
		t.Fatal("source version literal missing")
	}
	changed := strings.Replace(string(original), literal, `"9.8.7-dev"`, 1)
	changed = strings.ReplaceAll(changed, "CurrentVersion", "RenamedVersion")
	writeFixture(t, filepath.Join(root, sourcePath), []byte(changed))
	if err := project(root, false); err == nil {
		t.Fatal("stale projection accepted")
	}
	assertOutput(t, root, before)
	if err := project(root, true); err != nil {
		t.Fatal(err)
	}
	checkNativeVersion(t, root, "9.8.7-dev")
	after, err := os.ReadFile(filepath.Join(root, outputPath))
	if err != nil || bytes.Equal(before, after) {
		t.Fatal("changed source did not change generated library", err)
	}
	writeFixture(t, filepath.Join(root, sourcePath), []byte("invalid Gooo source"))
	if err := project(root, true); err == nil {
		t.Fatal("invalid source accepted")
	}
	assertOutput(t, root, after)
}

func projectionFixture(t *testing.T) (string, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("../..", sourcePath))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, sourcePath), raw)
	return root, raw
}

func writeFixture(t *testing.T, path string, raw []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func assertOutput(t *testing.T, root string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(root, outputPath))
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal("previous generated output changed", err)
	}
}

func checkNativeVersion(t *testing.T, root, want string) {
	t.Helper()
	writeFixture(t, filepath.Join(root, "go.mod"), []byte("module example.test/release\ngo 1.27.3\n"))
	source := "package releaseidentity\nimport \"testing\"\nfunc TestValue(t *testing.T) { if Evaluate() != " +
		"\"" + want + "\" { t.Fatal(Evaluate()) } }\n"
	writeFixture(t, filepath.Join(root, filepath.Dir(outputPath), "consumer_test.go"), []byte(source))
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-count=1", "./internal/meta/releaseidentity")
	command.Dir, command.Env = root, append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
}
