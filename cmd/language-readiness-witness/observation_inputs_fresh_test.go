package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestObservationInputPairFresh1890(t *testing.T) {
	for _, mode := range []string{"absent", "complete", "document-only", "policy-only"} {
		t.Run(mode, func(t *testing.T) {
			directory := t.TempDir()
			if mode == "complete" || mode == "document-only" {
				if err := os.WriteFile(filepath.Join(directory, "operation-observations.json"), []byte(`{"observations":[]}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "complete" || mode == "policy-only" {
				if err := os.WriteFile(filepath.Join(directory, "observation-policy.gooo"), []byte("package source1890\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			input, err := readObservationInputs(directory)
			if mode == "document-only" || mode == "policy-only" {
				if err == nil {
					t.Fatal("partial observation input silently fell back to legacy mode")
				}
				return
			}
			if err != nil || (mode == "absent" && input != nil) {
				t.Fatal(mode, input, err)
			}
			if mode == "complete" && (input == nil || !bytes.Equal(input.PolicySource, []byte("package source1890\n"))) {
				t.Fatal("observation source bytes changed")
			}
		})
	}
}
