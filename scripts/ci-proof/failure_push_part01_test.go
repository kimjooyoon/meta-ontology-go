package main

import (
	"strings"
	"testing"
)

func TestPullRequestPromotionFailureManifestUsesExactDevOwner(t *testing.T) {
	binding := validFailureBinding()
	binding.BaseRef = "main"
	binding.EventRef = "refs/pull/163/merge"
	binding.PRNumber = 163
	binding.OwnerBranch = "dev"
	input := validFailureInput()
	input.OwnerBranch = "dev"
	manifest, err := buildFailureManifest(input, binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Scope != "pr" || manifest.BaseRef != "main" || manifest.OwnerBranch != "dev" {
		t.Fatalf("exact dev-to-main promotion owner was not preserved: %+v", manifest)
	}
}
func TestPullRequestOwnerBindingRejectsInvalidPromotionRoutes(t *testing.T) {
	for name, route := range map[string]struct{ base, owner string }{
		"agent to main":     {base: "main", owner: "agent/ci-workflow"},
		"main to main":      {base: "main", owner: "main"},
		"dev to dev":        {base: "dev", owner: "dev"},
		"unknown base":      {base: "release", owner: "agent/ci-workflow"},
		"registry mismatch": {base: "dev", owner: "agent/not-registered"},
		"malformed owner":   {base: "main", owner: ""},
	} {
		t.Run(name, func(t *testing.T) {
			binding := validFailureBinding()
			binding.BaseRef = route.base
			binding.EventRef = "refs/pull/163/merge"
			binding.PRNumber = 163
			binding.OwnerBranch = route.owner
			err := validateFailureOwnerBinding(binding)
			if err == nil {
				t.Fatal("invalid pull-request owner route was accepted")
			}
			if name != "registry mismatch" && name != "malformed owner" && !strings.Contains(err.Error(), promotionOwnerBindingCode) {
				t.Fatalf("invalid promotion route did not report %s: %v", promotionOwnerBindingCode, err)
			}
		})
	}
}
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
