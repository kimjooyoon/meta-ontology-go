package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
	"testing"
)

func TestCISCOPE002MalformedAgentBranchFailsClosed(t *testing.T) {
	for _, branch := range []string{"agent/*", "agent/../ci-workflow", "agent//ci-workflow"} {
		if err := checkAgentPushBranch(branch); err == nil {
			t.Errorf("malformed agent branch %s was accepted", branch)
		}
	}
	if err := verify.CheckPathScopeForBranch(nil, "agent/*"); err == nil {
		t.Fatal("wildcard branch key was accepted")
	}
}
func TestCapModesRejectAmbiguousInvocation(t *testing.T) {
	if err := validateCapMode(true, true); err == nil {
		t.Fatal("caps-only and skip-caps were accepted together")
	}
	if err := validateCapMode(true, false); err != nil {
		t.Fatal(err)
	}
	if err := validateCapMode(false, true); err != nil {
		t.Fatal(err)
	}
}
