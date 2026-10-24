package moduleboundary

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryAliasesDoNotHideModuleCrossing(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"shared", "nested", "nested/own"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range map[string]string{
		"go.mod": "module example.test/root\n", "nested/go.mod": "module example.test/nested\n",
		"shared/source.go": "package p\n", "nested/own/source.go": "package p\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("../shared", filepath.Join(root, "nested", "cross")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("own", filepath.Join(root, "nested", "same")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("..", filepath.Join(root, "nested", "root")); err != nil {
		t.Fatal(err)
	}
	if err := CheckWriteSet(root, filepath.Join(root, "nested", "cross", "source.go")); err == nil {
		t.Fatal("directory link hid a cross-module source")
	}
	if err := CheckWriteSet(root, filepath.Join(root, "nested", "same", "source.go")); err != nil {
		t.Fatalf("same-module alias was rejected: %v", err)
	}
	if err := CheckWriteSet(root, filepath.Join(root, "nested", "root", "go.mod")); err == nil {
		t.Fatal("directory link to the repository root hid a module crossing")
	}
}
