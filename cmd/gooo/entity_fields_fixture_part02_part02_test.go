package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func filesystemDigest(t *testing.T, root string) string {
	t.Helper()
	var records []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		digest := ""
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			digest = semantic.StableHash(data)
		}
		records = append(records, fmt.Sprintf("%s:%o:%d:%s", filepath.ToSlash(relative), info.Mode().Perm(), info.Size(), digest))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(records, "\n")
}
