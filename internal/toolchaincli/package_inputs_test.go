package toolchaincli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedTestPackageObservesExactInputs(t *testing.T) {
	for _, mutation := range []string{"unchanged", "module", "extra-go", "missing", "symlink", "directory"} {
		t.Run(mutation, func(t *testing.T) {
			root := t.TempDir()
			contents := [3][]byte{[]byte("package sample"), []byte("module sample"), []byte("package sample")}
			for i, name := range GeneratedTestPackageNames() {
				if err := os.WriteFile(filepath.Join(root, name), contents[i], 0o600); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(root, "go.mod")
			var err error
			switch mutation {
			case "module":
				err = os.WriteFile(path, []byte("module other"), 0o600)
			case "extra-go":
				err = os.WriteFile(filepath.Join(root, "hidden.go"), []byte("package sample"), 0o600)
			case "missing":
				err = os.Remove(path)
			case "symlink", "directory":
				if err = os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if mutation == "symlink" {
					err = os.Symlink("semantic.gooo.go", path)
				} else {
					err = os.Mkdir(path, 0o700)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			err = VerifyGeneratedTestPackage(root, contents)
			if (err == nil) != (mutation == "unchanged") {
				t.Fatal("unexpected package boundary", mutation, err)
			}
		})
	}
}
