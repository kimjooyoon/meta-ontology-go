package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreviousObservationIsPreservedButNotActive(t *testing.T) {
	active := filepath.Join(t.TempDir(), "observations.json")
	if err := os.WriteFile(active, []byte("previous terminal evidence"), 0o600); err != nil {
		t.Fatal(err)
	}
	archived, err := archivePreviousObservation(active)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(active); !os.IsNotExist(err) {
		t.Fatalf("interrupted attempt could read old active evidence: %v", err)
	}
	data, err := os.ReadFile(archived)
	if err != nil || string(data) != "previous terminal evidence" {
		t.Fatalf("lost historical evidence: %q %v", data, err)
	}
}

func TestMissingPreviousObservationIsNotAnError(t *testing.T) {
	name, err := archivePreviousObservation(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil || name != "" {
		t.Fatalf("new output path: %q %v", name, err)
	}
}

func TestPreviousObservationDirectoryIsRejected(t *testing.T) {
	root := t.TempDir()
	if _, err := archivePreviousObservation(root); err == nil {
		t.Fatal("directory accepted as observation evidence")
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		t.Fatalf("rejected directory changed: %v", err)
	}
}
