package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunAnalyzeIsReadOnly(t *testing.T) {
	directory := t.TempDir()
	filename := filepath.Join(directory, "input.gooo")
	if err := os.WriteFile(filename, []byte(sourceOrderA), 0o640); err != nil {
		t.Fatal(err)
	}
	beforeBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeEntries := directoryEntries(t, directory)
	output, code, stderr := runAnalyzeFile(filename)
	if code != exitOK || len(stderr) != 0 || len(output) == 0 {
		t.Fatalf("read-only analyze = code %d, stderr=%q, output=%d bytes", code, stderr, len(output))
	}
	afterBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeBytes, afterBytes) || !os.SameFile(beforeInfo, afterInfo) || beforeInfo.Mode() != afterInfo.Mode() || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("analyze changed its input file")
	}
	if afterEntries := directoryEntries(t, directory); !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Fatalf("analyze changed directory entries: before=%v after=%v", beforeEntries, afterEntries)
	}
}

func TestRunVersionDoesNotWriteFilesystem(t *testing.T) {
	directory := t.TempDir()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
	before, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if code := runVersion([]string{"--json"}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("version code = %d", code)
	}
	after, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("version changed filesystem: before=%v after=%v", before, after)
	}
}
