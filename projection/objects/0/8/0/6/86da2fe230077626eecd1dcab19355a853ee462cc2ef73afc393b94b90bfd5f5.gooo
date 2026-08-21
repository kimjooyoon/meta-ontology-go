package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNoFailureClosureBindsHealthOnlyTuple(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildClosureManifest(validClosureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Status != "NO_TERMINAL_FAILURE" || manifest.Decision != "HEALTH_PASS_ONLY" || manifest.WriteEffect != "none" || len(manifest.TerminalFailures) != 0 {
		t.Fatalf("unexpected no-failure closure: %+v", manifest)
	}
	if err := validateClosureManifest(manifest, binding); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(manifest)
	if err != nil || !strings.Contains(string(encoded), `"schema":"gooo/ci-closure/v1"`) {
		t.Fatalf("closure schema missing: %s", encoded)
	}
}
func TestNoFailureClosureRejectsTerminalFailureData(t *testing.T) {
	input := validClosureInput()
	input.TerminalFailures = []failureJob{{ID: 12}}
	if _, err := buildClosureManifest(input, validFailureBinding()); err == nil {
		t.Fatal("no-failure closure accepted terminal failure data")
	}
}
func TestNoFailureClosureSupportsExactDevToMainPromotion(t *testing.T) {
	binding := validFailureBinding()
	binding.BaseRef = "main"
	binding.EventRef = "refs/pull/163/merge"
	binding.PRNumber = 163
	binding.OwnerBranch = "dev"
	manifest, err := buildClosureManifest(validClosureInputFor(binding), binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.BaseRef != "main" || manifest.OwnerBranch != "dev" || manifest.Decision != "HEALTH_PASS_ONLY" {
		t.Fatalf("exact dev-to-main promotion closure was not preserved: %+v", manifest)
	}
	if err := validateClosureManifest(manifest, binding); err != nil {
		t.Fatal(err)
	}
}
func TestNoFailureClosureAcceptsRegisteredDocsOwner(t *testing.T) {
	binding := validFailureBinding()
	binding.OwnerBranch = "agent/docs"
	manifest, err := buildClosureManifest(validClosureInputFor(binding), binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.OwnerBranch != binding.OwnerBranch || manifest.OwnerRef == "" {
		t.Fatalf("registered docs owner was not bound: %+v", manifest)
	}
	if err := validateClosureManifest(manifest, binding); err != nil {
		t.Fatal(err)
	}
}
func TestNoFailureClosureRejectsStaleCanonicalJob(t *testing.T) {
	input := validClosureInput()
	input.CanonicalJobs[0].HeadSHA = strings.Repeat("b", 40)
	if _, err := buildClosureManifest(input, validFailureBinding()); err == nil {
		t.Fatal("stale canonical job was accepted")
	}
}
