package main

import "testing"

func TestProtectedPushFailureManifestUsesExactBranchOwner(t *testing.T) {
	for _, branch := range []string{"dev", "main"} {
		t.Run(branch, func(t *testing.T) {
			binding := validFailureBinding()
			binding.Event = "push"
			binding.EventRef = "refs/heads/" + branch
			binding.BaseRef = branch
			binding.PRNumber = 0
			binding.OwnerBranch = branch
			input := validFailureInput()
			input.OwnerBranch = branch
			manifest, err := buildFailureManifest(input, binding)
			if err != nil {
				t.Fatal(err)
			}
			if manifest.Scope != branch || manifest.OwnerBranch != branch || manifest.EventRef != "refs/heads/"+branch {
				t.Fatalf("protected push identity was not preserved: %+v", manifest)
			}
		})
	}
}

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
