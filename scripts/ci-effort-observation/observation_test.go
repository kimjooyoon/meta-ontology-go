package main

import (
	"errors"
	"strings"
	"testing"
)

func TestDurationUsesPositiveIntegerMilliseconds(t *testing.T) {
	value, err := durationMS("2026-08-30T00:00:00.000000001Z", "2026-08-30T00:00:00.000000002Z")
	if err != nil || value != 1 {
		t.Fatalf("duration = %d, err = %v", value, err)
	}
}

func TestReuseRequiresEveryContextDigest(t *testing.T) {
	base := ReuseKey{HeadSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", InputDigest: "input", ToolchainDigest: "toolchain", CommandContextDigest: "command", EnvironmentAllowlistDigest: "environment", DependencyGraphDigest: "dependency", ExpectedResultDigest: "expected", OpenTofuReleaseDigest: "release"}
	mutations := []func(*ReuseKey){
		func(key *ReuseKey) { key.HeadSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" },
		func(key *ReuseKey) { key.InputDigest = "changed" },
		func(key *ReuseKey) { key.ToolchainDigest = "changed" },
		func(key *ReuseKey) { key.CommandContextDigest = "changed" },
		func(key *ReuseKey) { key.EnvironmentAllowlistDigest = "changed" },
		func(key *ReuseKey) { key.DependencyGraphDigest = "changed" },
		func(key *ReuseKey) { key.ExpectedResultDigest = "changed" },
		func(key *ReuseKey) { key.OpenTofuReleaseDigest = "changed" },
	}
	for index, mutate := range mutations {
		candidate := base
		mutate(&candidate)
		if sameReuseKey(base, candidate) {
			t.Fatalf("mutation %d was accepted", index)
		}
	}
}

func TestMissingPriorHasCompleteUnknownContext(t *testing.T) {
	unknown := priorMissingUnknown()
	if !validUnknown(unknown) || unknown.UnknownClass != "DIRECT_MISSING" || len(unknown.BlockedBy) != 0 {
		t.Fatalf("unknown context is incomplete: %+v", unknown)
	}
}

func TestCounterexamplesPreserveReuseBoundary(t *testing.T) {
	cases := fixedCounterexamples()
	if len(cases) != 5 || cases[0].Decision != "REFUTED" || cases[2].Unknown == nil {
		t.Fatalf("counterexamples = %+v", cases)
	}
}

func TestWorkflowCommandBindingRejectsSourceCommandDrift(t *testing.T) {
	source := []byte("jobs:\n  check:\n    name: check\n    steps:\n      - name: Verify\n        run: go test ./...\n")
	spec := OperationSpec{ID: "check", JobName: "check", StepName: "Verify", Command: []string{"go", "test", "./..."}}
	if _, err := bindWorkflowCommand(source, ".github/workflows/ci.yml", spec); err != nil {
		t.Fatalf("valid workflow command rejected: %v", err)
	}
	mutated := []byte(strings.ReplaceAll(string(source), "go test ./...", "go vet ./..."))
	if _, err := bindWorkflowCommand(mutated, ".github/workflows/ci.yml", spec); err == nil {
		t.Fatal("workflow command drift was accepted")
	}
}

func TestMissingCommandContextIsTypedUnknown(t *testing.T) {
	spec := OperationSpec{ID: "check", JobName: "check", StepName: "Verify", Kind: "VERIFICATION", Command: []string{"go", "test", "./..."}, ProofObligationID: "ci-effort/check"}
	operation := observeOperation(spec, nil, ".github/workflows/ci.yml", nil, errors.New("workflow source unavailable"))
	if operation.State != "UNKNOWN" || !validUnknown(operation.Unknown) || operation.Unknown.BlockedBy == nil {
		t.Fatalf("unknown operation lost causal context: %+v", operation)
	}
}

func TestMissingOpenTofuEvidenceIsTypedUnknown(t *testing.T) {
	value, err := observeOpenTofu("", "", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil || value.ArtifactID != 0 || !validUnknown(value.Unknown) {
		t.Fatalf("missing OpenTofu evidence = %+v, err = %v", value, err)
	}
}

func TestRepositoryMutationIsNotAValidObservation(t *testing.T) {
	decision, resolution, reason := classifyReport(Report{SourceRunConclusion: "success", RepositoryWrites: 1})
	if decision != "REFUTED" || resolution != "EXACT" || reason != "KNOWN_VERIFICATION_CONTRADICTION" {
		t.Fatalf("repository mutation classified as %s/%s/%s", decision, resolution, reason)
	}
}
