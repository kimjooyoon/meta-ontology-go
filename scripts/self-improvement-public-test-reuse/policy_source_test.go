package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestCanonicalPolicySourceBinding(t *testing.T) {
	fixture, err := os.ReadFile("../../examples/self-improvement-discovery/project.gooo")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"exact", "different-source", "missing-source"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			input := executionInput{Policy: filepath.Join(root, "policy.gooo"), Source: filepath.Join(root, "source.gooo")}
			if err := os.WriteFile(input.Policy, fixture, 0o600); err != nil {
				t.Fatal(err)
			}
			source := append([]byte(nil), fixture...)
			if name == "different-source" {
				source = append(source, []byte("\n# a different source observation\n")...)
			}
			if name != "missing-source" {
				if err := os.WriteFile(input.Source, source, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			policy, policyBytes, sourceBytes, err := readCanonicalPolicy(input)
			if name != "exact" {
				wanted := "same frozen .gooo file"
				if name == "missing-source" {
					wanted = "read canonical source"
				}
				if err == nil || !strings.Contains(err.Error(), wanted) || policyBytes != nil || sourceBytes != nil {
					t.Fatalf("unbound source was retained: %v", err)
				}
				return
			}
			if err != nil || !bytes.Equal(policyBytes, fixture) || !bytes.Equal(sourceBytes, fixture) ||
				policy.SourceDigest != cache.HashBytes(fixture).String() {
				t.Fatal("exact source observations were not preserved", err)
			}
			sourceBytes[0] ^= 1
			if !bytes.Equal(policyBytes, fixture) {
				t.Fatal("distinct file observations share mutable bytes")
			}
		})
	}
}
