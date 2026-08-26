package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRefactorFailsClosedForUnsafeSources(t *testing.T) {
	tests := map[string]string{
		"parse": "package broken\nfunc A( {\nfunc B() {}\n",
		"init":  "package sample\nfunc init() {}\nfunc A() {}\nfunc B() {}\n",
		"var":   "package sample\nvar state = 1\nfunc A() {}\nfunc B() {}\n",
		"cgo":   "package sample\nimport \"C\"\nfunc A() {}\nfunc B() {}\n",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "source.go")
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := refactor(path, options{maxLines: 3, maxEntries: 10, write: true}); err == nil {
				t.Fatal("refactor should fail closed")
			}
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("original source was not preserved: %v", err)
			}
		})
	}
}
