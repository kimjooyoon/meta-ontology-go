package main

import "testing"

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
