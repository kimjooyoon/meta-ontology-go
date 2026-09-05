package main

import "testing"

func TestInvocationIDSeparatesJobAttemptsAndOutputs(t *testing.T) {
	t.Setenv("GITHUB_RUN_ID", "42")
	t.Setenv("GITHUB_RUN_ATTEMPT", "1")
	t.Setenv("GITHUB_JOB", "focused")
	first := invocationID("/runner/first/ledger.json")
	replay := invocationID("/runner/replay/ledger.json")
	if first == replay {
		t.Fatalf("same-job output invocations were combined: %q", first)
	}

	t.Setenv("GITHUB_RUN_ATTEMPT", "2")
	retry := invocationID("/runner/first/ledger.json")
	if first == retry {
		t.Fatalf("different attempts were combined: %q", first)
	}
}
