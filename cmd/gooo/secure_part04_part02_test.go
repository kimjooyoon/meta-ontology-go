package main

import (
	"os"
	"path/filepath"
	"testing"
)

// CLI-ATOMIC-001: successful output leaves no temporary sibling.
func TestCLIATOMIC001WritesAndCleansTemporaryOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, generatedFileName)
	if err := writeAtomicFile(path, []byte("atomic bytes\n")); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != generatedFileName {
		t.Fatalf("temporary output was not cleaned: %+v", entries)
	}
}
