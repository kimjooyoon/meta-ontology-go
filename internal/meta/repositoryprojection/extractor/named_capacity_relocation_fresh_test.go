package extractor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNamedCapacityRelocation21107(t *testing.T) {
	targets := []string{"sample/one#Alpha21107", "sample/one#Reader21107.Measure21107"}
	layouts := []map[string]string{
		{"original.go": "package sample\ntype Reader21107 struct{}\nfunc Alpha21107()int{return 21107}\nfunc(Reader21107)Measure21107()int{return 17}\n"},
		{"moved_a.go": "package sample\nfunc Alpha21107()int{return 21107}\n", "moved_b.go": "package sample\ntype Reader21107 struct{}\nfunc(Reader21107)Measure21107()int{return 17}\n", "ignored_test.go": "package sample\nfunc Alpha21107()int{return -1}\n"},
	}
	for _, files := range layouts {
		root := t.TempDir()
		directory := filepath.Join(root, "sample", "one")
		if err := os.MkdirAll(directory, 0700); err != nil {
			t.Fatal(err)
		}
		for path, content := range files {
			if err := os.WriteFile(filepath.Join(directory, path), []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
		}
		requireNamedCapacities(t, root, targets)
		if _, err := collectNamedCapacities(root, append(append([]string{}, targets...), "sample/one#Missing21107")); err == nil || !strings.Contains(err.Error(), "missing") {
			t.Fatal("missing declaration was silently removed", err)
		}
		if err := os.WriteFile(filepath.Join(directory, "duplicate.go"), []byte("package sample\nfunc Alpha21107()int{return 29}\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := collectNamedCapacities(root, targets); err == nil || !strings.Contains(err.Error(), "duplicate") {
			t.Fatal("ambiguous declaration accepted", err)
		}
	}
	t.Log("new two-declaration source layouts preserve package/receiver identities across file splits;missing/duplicate rejected;producer/model0;historical source cohorts not executed")
}
