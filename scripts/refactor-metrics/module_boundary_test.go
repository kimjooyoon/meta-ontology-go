package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestDirectRefactorHonorsOwningModule(t *testing.T) {
	for _, kind := range []string{"nested", "linked", "newer"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "nested"), 0o700); err != nil {
				t.Fatal(err)
			}
			version := "1.27.3"
			if kind == "newer" {
				version = "1.99.0"
			}
			source := "package p\n\nfunc value() int {\n x := 1\n return x\n}\n"
			for name, content := range map[string]string{
				"go.mod": "module example.test/root\n", "root.go": source,
				"nested/go.mod": "module example.test/nested\ngo " + version + "\n",
			} {
				if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(root, "nested", "source.go")
			var err error
			if kind == "linked" {
				err = os.Symlink("../root.go", path)
			} else {
				err = os.WriteFile(path, []byte(source), 0o600)
			}
			if err != nil {
				t.Fatal(err)
			}
			subject := sourcepolicy.SourceSubject{Path: "nested/source.go", Line: 3, Name: "value"}
			err = transformCandidate(root, candidate{raw: subject.String(), source: subject}, true)
			got, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if kind == "nested" {
				if err != nil || !strings.Contains(string(got), "return 1") {
					t.Fatalf("nested refactor failed: %v %s", err, got)
				}
			} else if err == nil || string(got) != source {
				t.Fatalf("unsupported module was modified: %v %s", err, got)
			}
		})
	}
}
