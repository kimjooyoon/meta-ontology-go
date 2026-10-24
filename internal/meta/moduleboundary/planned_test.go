package moduleboundary

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlannedPackagePathsRetainModuleWithoutCreatingDirectories(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"nested/pkg", "nested/shared", "outside"} {
		if err := os.MkdirAll(filepath.Join(root, path), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for path, content := range map[string]string{
		"go.mod": "module example.test/root\n", "nested/go.mod": "module example.test/nested\n",
	} {
		if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("../shared", filepath.Join(root, "nested/pkg/alias")); err != nil {
		t.Skip(err)
	}
	subject := filepath.Join(root, "nested/pkg")
	path := filepath.Join(subject, "alias/new/deeper/source.go")
	owner, err := CheckPackageWriteSet(root, subject, path)
	if err != nil || owner != "nested" {
		t.Fatalf("same-module planned path: owner=%q error=%v", owner, err)
	}
	if _, err := os.Stat(filepath.Join(root, "nested/shared/new")); !os.IsNotExist(err) {
		t.Fatalf("planning created a directory: %v", err)
	}
	for _, name := range []string{"go.mod", "go.sum", "go.work", "go.work.sum"} {
		if _, err := CheckPackageWriteSet(root, subject, filepath.Join(subject, "new", name)); err == nil {
			t.Fatalf("package recipe accepted module graph write: %s", name)
		}
	}
}
