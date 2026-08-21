package main

import (
	"path/filepath"
	"testing"
)

func TestGeneratedPathPreservesGoIdentity(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sample.go", "sample_part01.go"},
		{"sample_test.go", "sample_part01_test.go"},
		{"sample_linux.go", "sample_part01_linux.go"},
		{"sample_amd64.go", "sample_part01_amd64.go"},
		{"sample_linux_amd64.go", "sample_part01_linux_amd64.go"},
		{"sample_linux_amd64_test.go", "sample_part01_linux_amd64_test.go"},
		{"sample_linux_test.gooo", "sample_part01_linux_test.gooo"},
	}
	for _, test := range tests {
		got, err := generatedPath(filepath.Join("root", test.input), 1)
		if err != nil {
			t.Fatalf("generatedPath(%q): %v", test.input, err)
		}
		if base := filepath.Base(got); base != test.want {
			t.Errorf("generatedPath(%q) = %q, want %q", test.input, base, test.want)
		}
	}
}

func TestLineCountMatchesSourcePolicySemantics(t *testing.T) {
	for source, want := range map[string]int{"": 0, "a": 1, "a\n": 1, "a\nb": 2, "a\nb\n": 2} {
		if got := lineCount([]byte(source)); got != want {
			t.Errorf("lineCount(%q) = %d, want %d", source, got, want)
		}
	}
}
