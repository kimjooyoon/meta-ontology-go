package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitTemporaryFailureRetainsOriginal(t *testing.T) {
	for _, failedAt := range []int{0, 1, 2} {
		t.Run(fmt.Sprint(failedAt), func(t *testing.T) {
			root := t.TempDir()
			before := []byte("package fixture\nfunc First() {}\nfunc Second() {}\n")
			plan := splitPlan{Directory: root, Mode: 0o600, Parts: []splitPart{
				{Path: filepath.Join(root, "source.go"), Data: []byte("package fixture\nfunc First() {}\n")},
				{Path: filepath.Join(root, "second.go"), Data: []byte("package fixture\nfunc Second() {}\n")},
			}}
			if err := os.WriteFile(plan.Parts[0].Path, before, 0o600); err != nil {
				t.Fatal(err)
			}
			calls := 0
			create := func(dir, pattern string) (*os.File, error) {
				calls++
				if calls == failedAt {
					return splitReadOnlyTemp(dir, pattern)
				}
				return os.CreateTemp(dir, pattern)
			}
			var events []splitEvent
			err := applySplitWithTemp(plan, func(event splitEvent) { events = append(events, event) }, create)
			if (err == nil) != (failedAt == 0) {
				t.Error("temporary failure lost its error", err)
			}
			if failedAt != 0 {
				var operation *os.PathError
				if !errors.As(err, &operation) || operation.Op != "write" {
					t.Error("write cause lost", err)
				}
				got, readErr := os.ReadFile(plan.Parts[0].Path)
				if readErr != nil || !bytes.Equal(got, before) {
					t.Errorf("original source changed: %q (%v)", got, readErr)
				}
				for _, event := range events {
					if event.Kind == "RENAME_CREATE" || event.Kind == "RENAME_REPLACE" {
						t.Error("failed preparation reached publication", event)
					}
				}
			}
			checkSplitStagingFiles(t, root, plan, failedAt == 0)
		})
	}
}

func splitReadOnlyTemp(directory, pattern string) (*os.File, error) {
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

func checkSplitStagingFiles(t *testing.T, root string, plan splitPlan, success bool) {
	t.Helper()
	expectedFiles := 1
	if success {
		expectedFiles = 2
		for _, part := range plan.Parts {
			got, err := os.ReadFile(part.Path)
			if err != nil || !bytes.Equal(got, part.Data) {
				t.Error("successful partition changed", part.Path, err)
			}
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != expectedFiles {
		t.Error("unexpected published or temporary file", entries, err)
	}
}
