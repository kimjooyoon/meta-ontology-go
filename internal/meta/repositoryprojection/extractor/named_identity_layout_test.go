package extractor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNamedCapacityIdentityAfterFileMove(t *testing.T) {
	root := t.TempDir()
	write := func(name, source string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	// A stub and relocated test declarations mirror the post-projection layout.
	write("reader_test.go", "package fixture\n")
	write("reader_part07_test.go", "package fixture\nfunc ReadOrigin() int { return 1 }\n")
	write("reader_part21_test.go", "package fixture\nfunc BindOrigin() int { return 2 }\n")
	targets := []string{".#ReadOrigin", ".#BindOrigin"}
	observed, err := collectNamedCapacityFiles(root, targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
	if err := os.Rename(filepath.Join(root, "reader_part07_test.go"), filepath.Join(root, "another_test.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := collectNamedCapacityFiles(root, targets, true); err != nil {
		t.Fatal(err)
	}
	write("duplicate_test.go", "package fixture\nfunc ReadOrigin() int { return 3 }\n")
	if _, err := collectNamedCapacityFiles(root, targets, true); err == nil {
		t.Fatal("duplicate declaration accepted")
	}
	if err := os.Remove(filepath.Join(root, "duplicate_test.go")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "another_test.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := collectNamedCapacityFiles(root, targets, true); err == nil {
		t.Fatal("missing declaration accepted")
	}
}
