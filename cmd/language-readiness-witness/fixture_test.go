package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func conceptInput(t *testing.T, root string) string {
	t.Helper()
	data, err := json.MarshalIndent(languageconcept.BuildArtifact(os.DirFS(root)), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "language-concept-artifact.json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
