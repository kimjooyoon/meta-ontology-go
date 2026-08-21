package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCLIPATH004CommandRejectsSymlinkRootWithoutMutation(t *testing.T) {
	actual := t.TempDir()
	parent := t.TempDir()
	link := filepath.Join(parent, "out")
	if err := os.Symlink(actual, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	sentinel := filepath.Join(actual, "sentinel.go")
	original := []byte("sentinel bytes\n")
	if err := os.WriteFile(sentinel, original, 0o640); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(sentinel)
	if err != nil {
		t.Fatal(err)
	}
	beforeActual := directoryEntries(t, actual)
	beforeParent := directoryEntries(t, parent)
	run := func() (int, string, string) {
		var stdout, stderr bytes.Buffer
		code := runGenerate([]string{"billing.gooo", "--out", link}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
	firstCode, firstOut, firstErr := run()
	secondCode, secondOut, secondErr := run()
	if firstCode != exitFailure || secondCode != exitFailure || firstOut != "" || secondOut != "" || firstErr != secondErr || !strings.Contains(firstErr, "symbolic link") {
		t.Fatalf("symlink root rejection changed: first=%d/%q/%q second=%d/%q/%q", firstCode, firstOut, firstErr, secondCode, secondOut, secondErr)
	}
	after, err := os.Stat(sentinel)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) || !reflect.DeepEqual(beforeActual, directoryEntries(t, actual)) || !reflect.DeepEqual(beforeParent, directoryEntries(t, parent)) {
		t.Fatal("symlink root rejection mutated filesystem state")
	}
}
