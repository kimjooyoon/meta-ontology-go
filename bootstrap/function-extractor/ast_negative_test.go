package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGenericExtractionRejectsUnsafeInputs(t *testing.T) {
	cases := []struct{ name, logical, source, reason string }{
		{"unsafe path", "../escape.go", "package p\nfunc F() {}\n", "WRITE_SET_ESCAPE"},
		{"unresolved import", "x.go", "package p\nimport \"example.invalid/missing\"\nfunc F() {}\n", "UNRESOLVED_IMPORT"},
		{"build conflict", "x.go", "//go:build linux\n// +build windows\n\npackage p\nfunc F() {}\n", "BUILD_TAG_CONFLICT"},
		{"identity collision", "x.go", "package p\nfunc F() {}\nfunc F() {}\n", "DECLARATION_IDENTITY_COLLISION"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if tc.logical == "x.go" {
				if err := os.WriteFile(filepath.Join(root, tc.logical), []byte(tc.source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			_, _, err := genericASTExtraction(root, tc.logical)
			var failure extractionFailure
			if !errors.As(err, &failure) || failure.Reason != tc.reason {
				t.Fatalf("reason=%v want %s", err, tc.reason)
			}
		})
	}
}
