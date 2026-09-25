package main

import "testing"

func TestFailureManifestPreservesTypedTerminalFailureEvidence(t *testing.T) {
	input := validFailureInput()
	input.TerminalFailureEvidence = []terminalFailureEvidence{{
		Job: input.Job, Code: "CI-TEST-001", Classification: "known_job", Reason: "catalog_mapping:go test",
	}}
	manifest, err := buildFailureManifest(input, validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	if err := validateFailureManifest(manifest, validFailureBinding()); err != nil {
		t.Fatal(err)
	}
	manifest.TerminalFailureEvidence[0].Classification = "unclassified_job"
	if err := validateFailureManifest(manifest, validFailureBinding()); err == nil {
		t.Fatal("tampered terminal failure classification was accepted")
	}
}

func TestFailureManifestRequiresUnclassifiedCodeForUnknownTerminalEvidence(t *testing.T) {
	input := validFailureInput()
	input.TerminalFailures[0].Name = "future terminal check"
	input.TerminalFailureEvidence = []terminalFailureEvidence{{
		Job: input.TerminalFailures[0], Code: "CI-TEST-001", Classification: "unclassified_job", Reason: "unknown_terminal_job",
	}}
	if _, err := buildFailureManifest(input, validFailureBinding()); err == nil {
		t.Fatal("unknown terminal evidence was accepted with a known failure code")
	}
}
