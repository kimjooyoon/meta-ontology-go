package main

import (
	"testing"
)

func TestProtectedPushFailureManifestRejectsUnknownOrStaleOwner(t *testing.T) {
	for name, mutate := range map[string]func(*failureBinding){
		"unknown protected branch": func(binding *failureBinding) {
			binding.BaseRef = "release"
			binding.EventRef = "refs/heads/release"
			binding.HeadBranch = "release"
		},
		"stale owner": func(binding *failureBinding) {
			binding.BaseRef = "main"
			binding.EventRef = "refs/heads/main"
			binding.HeadBranch = "integration"
		},
		"retired protected branch": func(binding *failureBinding) {
			binding.BaseRef = "integration"
			binding.EventRef = "refs/heads/integration"
			binding.HeadBranch = "integration"
		},
		"pull request number": func(binding *failureBinding) {
			binding.BaseRef = "dev"
			binding.EventRef = "refs/heads/dev"
			binding.HeadBranch = "dev"
			binding.PRNumber = 105
		},
	} {
		t.Run(name, func(t *testing.T) {
			binding := validFailureBinding()
			binding.Event = "push"
			mutate(&binding)
			input := validFailureInput()
			input.HeadBranch = binding.HeadBranch
			if _, err := buildFailureManifest(input, binding); err == nil {
				t.Fatal("invalid protected push owner was accepted")
			}
		})
	}
}
