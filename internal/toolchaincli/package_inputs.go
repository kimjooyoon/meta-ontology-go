package toolchaincli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func GeneratedTestPackageNames() [3]string {
	return [3]string{"generated_project_test.go", "go.mod", "semantic.gooo.go"}
}

// VerifyGeneratedTestPackage observes exact membership and selected bytes at
// this boundary. It does not establish an atomic snapshot or SDK closure.
func VerifyGeneratedTestPackage(root string, expected [3][]byte) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	names := GeneratedTestPackageNames()
	if len(entries) != len(names) {
		return fmt.Errorf("generated test package requires exactly %d input files", len(names))
	}
	for index, entry := range entries {
		if entry.Name() != names[index] {
			return fmt.Errorf("unexpected generated test package entry %q", entry.Name())
		}
		if err := verifyGeneratedTestFile(filepath.Join(root, entry.Name()), expected[index]); err != nil {
			return err
		}
	}
	return nil
}

func verifyGeneratedTestFile(path string, expected []byte) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("generated package input %q is not a regular file", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	actual, err := io.ReadAll(io.LimitReader(file, int64(len(expected))+1))
	if err != nil {
		return err
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("generated package input %q differs from selected bytes", path)
	}
	return nil
}
