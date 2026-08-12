package main

import (
	"strings"
	"testing"
)

func TestFailureManifestBuildsCanonicalPROVRelations(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != failureSchema || manifest.Version != 1 || manifest.Code != "CI-TEST-001" || manifest.Scope != "pr" || manifest.BlockingScope != "local" || !manifest.Parallelizable {
		t.Fatalf("unexpected failure manifest classification: %+v", manifest)
	}
	if manifest.Provenance.WasGeneratedBy != manifest.Activity || manifest.Provenance.WasAssociatedWith != manifest.Agent || len(manifest.Provenance.WasDerivedFrom) != 2 || len(manifest.Provenance.HadPrimarySource) != 2 {
		t.Fatal("PROV relations were not canonicalized")
	}
	if err := validateFailureManifest(manifest, binding); err != nil {
		t.Fatal(err)
	}
}

func TestFailureManifestRejectsStaleHead(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.HeadSHA = strings.Repeat("b", 40)
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("stale failure head was accepted")
	}
}

func TestFailureManifestRejectsCheckoutMismatch(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.CheckoutRef = strings.Repeat("b", 40)
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("checkout ref mismatch was accepted")
	}
}

func TestFailureManifestRejectsTamperedProvenance(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.Provenance.WasGeneratedBy = "urn:gooo:ci-run:replayed"
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered provenance relation was accepted")
	}
}

func TestFailureManifestRejectsCallerSuppliedTuple(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.RunID++
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("caller-supplied run tuple was accepted")
	}
}

func TestFailureManifestRejectsMismatchedJobBinding(t *testing.T) {
	binding := validFailureBinding()
	input := validFailureInput()
	input.Job.HeadSHA = strings.Repeat("b", 40)
	if _, err := buildFailureManifest(input, binding); err == nil {
		t.Fatal("mismatched failure job head was accepted")
	}
}

func TestFailureManifestRejectsUnknownBinding(t *testing.T) {
	binding := validFailureBinding()
	binding.Actor = "unknown"
	if _, err := buildFailureManifest(validFailureInput(), binding); err == nil {
		t.Fatal("unknown agent binding was accepted")
	}
}

func validFailureBinding() failureBinding {
	return failureBinding{
		Repository: "owner/repo", Event: "pull_request", EventRef: "refs/pull/7/merge", CheckoutRef: strings.Repeat("a", 40), BaseRef: "integration",
		BaseSHA: strings.Repeat("b", 40), HeadSHA: strings.Repeat("a", 40), WorkflowSHA: strings.Repeat("c", 40),
		PRNumber: 7, RunID: 9, RunAttempt: 2, Actor: "builder",
	}
}

func validFailureInput() failureInput {
	head := strings.Repeat("a", 40)
	return failureInput{
		Code: "CI-TEST-001", Message: "go test failed in the exact PR run", Remediation: "reproduce and fix the failing test",
		Job: failureJob{ID: 11, Name: "go test", Status: "completed", Conclusion: "failure", HeadSHA: head, RunID: 9, RunAttempt: 2},
	}
}
