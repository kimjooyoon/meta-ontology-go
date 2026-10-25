package sourcerepack

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func stagingFixture(t *testing.T) repackPlan {
	t.Helper()
	root := t.TempDir()
	plan := repackPlan{Edits: []fileEdit{
		{Path: filepath.Join(root, "source.go"), Subject: "source.go", Before: []byte("package fixture\nfunc First(){}\n"),
			After: []byte("package fixture\n"), Mode: 0o600},
		{Path: filepath.Join(root, "target.go"), Subject: "target.go", Before: []byte("package fixture\n"),
			After: []byte("package fixture\nfunc First(){}\n"), Mode: 0o600},
	}}
	for _, edit := range plan.Edits {
		if err := os.WriteFile(edit.Path, edit.Before, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return plan
}

func readOnlyTemporary(directory, pattern string) (*os.File, error) {
	file, err := os.CreateTemp(directory, pattern)
	if err != nil {
		return nil, err
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		return nil, err
	}
	return os.Open(name)
}

func TestStagingRetainsFirstFailedOperation(t *testing.T) {
	root := t.TempDir()
	createClosed := func(directory, pattern string) (*os.File, error) {
		file, err := os.CreateTemp(directory, pattern)
		if err == nil {
			err = file.Close()
		}
		return file, err
	}
	path, err := stageWithTemp(filepath.Join(root, "source.go"), []byte("package fixture\n"), 0o600, createClosed)
	var operation *os.PathError
	if path != "" || !errors.As(err, &operation) || operation.Op != "chmod" {
		t.Fatal("first failure was replaced by a later operation", path, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatal("failed stage retained its temporary file", entries, err)
	}
}
