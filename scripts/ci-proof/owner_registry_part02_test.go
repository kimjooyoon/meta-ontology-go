package main

import (
	"testing"
)

func TestFailureOwnerRegistryRejectsMalformedRows(t *testing.T) {
	tests := map[string]struct {
		branch string
		mutate func(*failureOwnerRegistry)
	}{
		"unknown owner":   {branch: "agent/not-registered"},
		"wildcard owner":  {branch: "agent/*"},
		"protected owner": {branch: "dev"},
		"wildcard registry row": {
			branch: "agent/docs",
			mutate: func(registry *failureOwnerRegistry) {
				registry.Ownership = append(registry.Ownership, failureOwnerRegistryEntry{Branch: "agent/*", Paths: []string{"docs/**"}})
			},
		},
		"protected registry row": {
			branch: "agent/docs",
			mutate: func(registry *failureOwnerRegistry) {
				registry.Ownership = append(registry.Ownership, failureOwnerRegistryEntry{Branch: "main", Paths: []string{"docs/**"}})
			},
		},
		"duplicate branch": {
			branch: "agent/docs",
			mutate: func(registry *failureOwnerRegistry) {
				registry.Ownership = append(registry.Ownership, registry.Ownership[0])
			},
		},
		"empty paths": {
			branch: "agent/docs",
			mutate: func(registry *failureOwnerRegistry) {
				registry.Ownership[0].Paths = nil
			},
		},
		"malformed path": {
			branch: "agent/docs",
			mutate: func(registry *failureOwnerRegistry) {
				registry.Ownership[0].Paths = []string{"docs/**/nested"}
			},
		},
		"duplicate path": {
			branch: "agent/docs",
			mutate: func(registry *failureOwnerRegistry) {
				registry.Ownership[0].Paths = []string{"internal/analyzer/**", "internal/analyzer/**"}
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data := failureOwnerRegistryDocument(t, test.mutate)
			if err := validateFailureOwnerRegistryDocument(data, test.branch); err == nil {
				t.Fatal("malformed or unauthorized owner registry was accepted")
			}
		})
	}
}
