package publicpartialreuse

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCompilerDigestRetainsNativeResourceReaderOrigins(t *testing.T) {
	root := t.TempDir()
	for _, name := range compilerManifestPaths {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	before, err := CompilerDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"model.go", "resource_darwin.go", "resource_linux.go", "resource_other.go"} {
		relative := "internal/toolchaincli/" + name
		if !slices.Contains(compilerManifestPaths, relative) {
			t.Fatal("native resource origin is missing", relative)
		}
		if err := os.WriteFile(filepath.Join(root, relative), []byte("changed "+relative), 0o600); err != nil {
			t.Fatal(err)
		}
		after, err := CompilerDigest(root)
		if err != nil || after == before {
			t.Fatal("resource-reader change retained an old compiler identity", relative, err)
		}
		before = after
	}
}
