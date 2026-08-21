package main

import (
	"testing"
)

func TestProtectedPushFailureManifestRejectsUnknownOrStaleOwner(t *testing.T) {
	for name, mutate := range map[string]func(*failureBinding){
		"unknown protected branch": func(binding *failureBinding) {
			binding.BaseRef = "release"
			binding.EventRef = "refs/heads/release"
			binding.OwnerBranch = "release"
		},
		"stale owner": func(binding *failureBinding) {
			binding.BaseRef = "main"
			binding.EventRef = "refs/heads/main"
			binding.OwnerBranch = "integration"
		},
		"retired protected branch": func(binding *failureBinding) {
			binding.BaseRef = "integration"
			binding.EventRef = "refs/heads/integration"
			binding.OwnerBranch = "integration"
		},
		"pull request number": func(binding *failureBinding) {
			binding.BaseRef = "dev"
			binding.EventRef = "refs/heads/dev"
			binding.OwnerBranch = "dev"
			binding.PRNumber = 105
		},
	} {
		t.Run(name, func(t *testing.T) {
			binding := validFailureBinding()
			binding.Event = "push"
			mutate(&binding)
			input := validFailureInput()
			input.OwnerBranch = binding.OwnerBranch
			if _, err := buildFailureManifest(input, binding); err == nil {
				t.Fatal("invalid protected push owner was accepted")
			}
		})
	}
}
