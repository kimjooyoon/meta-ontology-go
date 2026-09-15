package main

import "testing"

func TestInvocationIDSeparatesJobAttemptsAndOutputs(t *testing.T) {
	t.Setenv("GITHUB_RUN_ID", "42")
	t.Setenv("GITHUB_RUN_ATTEMPT", "1")
	t.Setenv("GITHUB_JOB", "focused")
	first, err := invocationID("/runner/first/ledger.json")
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := invocationID("/runner/first/ledger.json")
	if err != nil {
		t.Fatal(err)
	}
	if first == repeated {
		t.Fatalf("same-output invocations were combined: %q", first)
	}
	replay, err := invocationID("/runner/replay/ledger.json")
	if err != nil {
		t.Fatal(err)
	}
	if first == replay {
		t.Fatalf("same-job output invocations were combined: %q", first)
	}

	t.Setenv("GITHUB_RUN_ATTEMPT", "2")
	retry, err := invocationID("/runner/first/ledger.json")
	if err != nil {
		t.Fatal(err)
	}
	if first == retry {
		t.Fatalf("different attempts were combined: %q", first)
	}
}
