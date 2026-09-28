package main

import (
	"strings"
	"testing"
)

func TestPullRequestPromotionFailureManifestUsesExactDevHead(t *testing.T) {
	binding := validFailureBinding()
	binding.BaseRef = "main"
	binding.EventRef = "refs/pull/163/merge"
	binding.PRNumber = 163
	binding.HeadBranch = "dev"
	input := validFailureInput()
	input.HeadBranch = "dev"
	manifest, err := buildFailureManifest(input, binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Scope != "pr" || manifest.BaseRef != "main" || manifest.HeadBranch != "dev" {
		t.Fatalf("exact dev-to-main promotion owner was not preserved: %+v", manifest)
	}
}

func TestPullRequestPromotionFailureManifestPreservesSnapshotBranchIdentity(t *testing.T) {
	branch := "agent/main-promotion-snapshot-" + strings.Repeat("a", 40)
	binding := validFailureBinding()
	binding.BaseRef = "main"
	binding.EventRef = "refs/pull/163/merge"
	binding.PRNumber = 163
	binding.HeadBranch = branch
	input := validFailureInput()
	input.HeadBranch = branch
	manifest, err := buildFailureManifest(input, binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Scope != "pr" || manifest.BaseRef != "main" || manifest.HeadBranch != branch {
		t.Fatalf("snapshot promotion identity was not preserved: %+v", manifest)
	}
}
func TestPullRequestBranchBindingRejectsInvalidPromotionRoutes(t *testing.T) {
	for name, route := range map[string]struct{ base, head string }{
		"agent to main":      {base: "main", head: "agent/ci-workflow"},
		"main to main":       {base: "main", head: "main"},
		"dev to dev":         {base: "dev", head: "dev"},
		"unknown base":       {base: "release", head: "agent/ci-workflow"},
		"malformed head":     {base: "main", head: ""},
		"malformed snapshot": {base: "main", head: "agent/main-promotion-snapshot-short"},
	} {
		t.Run(name, func(t *testing.T) {
			binding := validFailureBinding()
			binding.BaseRef = route.base
			binding.EventRef = "refs/pull/163/merge"
			binding.PRNumber = 163
			binding.HeadBranch = route.head
			err := validateFailureBranchBinding(binding)
			if err == nil {
				t.Fatal("invalid pull-request owner route was accepted")
			}
			if name != "malformed head" && !strings.Contains(err.Error(), promotionBranchBindingCode) {
				t.Fatalf("invalid promotion route did not report %s: %v", promotionBranchBindingCode, err)
			}
		})
	}
}
func TestProtectedPushFailureManifestUsesExactBranchIdentity(t *testing.T) {
	for _, branch := range []string{"dev", "main"} {
		t.Run(branch, func(t *testing.T) {
			binding := validFailureBinding()
			binding.Event = "push"
			binding.EventRef = "refs/heads/" + branch
			binding.BaseRef = branch
			binding.PRNumber = 0
			binding.HeadBranch = branch
			input := validFailureInput()
			input.HeadBranch = branch
			manifest, err := buildFailureManifest(input, binding)
			if err != nil {
				t.Fatal(err)
			}
			if manifest.Scope != branch || manifest.HeadBranch != branch || manifest.EventRef != "refs/heads/"+branch {
				t.Fatalf("protected push identity was not preserved: %+v", manifest)
			}
		})
	}
}

func TestFailureManifestAcceptsUnregisteredFeatureBranch(t *testing.T) {
	binding := validFailureBinding()
	binding.HeadBranch = "agent/system-derived-scope"
	input := validFailureInput()
	input.HeadBranch = binding.HeadBranch
	manifest, err := buildFailureManifest(input, binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.HeadBranch != binding.HeadBranch {
		t.Fatalf("exact feature branch identity was not bound: %+v", manifest)
	}
}
