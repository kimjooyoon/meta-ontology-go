package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func projectionFiles(t *testing.T, directory string) map[string][]byte {
	t.Helper()
	const name = "registry.projection.json"
	raw, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := projectionbundle.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	result := map[string][]byte{name: raw}
	for _, member := range manifest.Members {
		content, err := os.ReadFile(filepath.Join(directory, member.Path))
		if err != nil || projectionbundle.Digest(content) != member.Digest {
			t.Fatal("projection manifest member differs:", member.Path, err)
		}
		result[member.Path] = content
	}
	return result
}
