package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectedPolicySeparatesSourceAndStorageTopology(t *testing.T) {
	source, storage := t.TempDir(), t.TempDir()
	for index := range 11 {
		name := filepath.Join(source, fmt.Sprintf("note-%02d.txt", index))
		if err := os.WriteFile(name, []byte("metric\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(source, "source.go"), []byte("package source\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	objects := filepath.Join(storage, "objects")
	if err := os.Mkdir(objects, 0o755); err != nil {
		t.Fatal(err)
	}
	policy := DefaultLinePolicy()
	if err := CheckProjectedSourcePolicy(source, storage, nil, policy); err != nil {
		t.Fatalf("logical topology was treated as physical: %v", err)
	}
	for index := range 11 {
		name := filepath.Join(objects, fmt.Sprintf("blob-%02d", index))
		if err := os.WriteFile(name, []byte("object\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := CheckProjectedSourcePolicy(source, storage, nil, policy); err == nil {
		t.Fatal("nonconforming physical projection was accepted")
	}
}
