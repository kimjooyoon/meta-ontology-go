package toolchaincli

import "testing"

func TestDeterministicReplayPreservesProcessContract(t *testing.T) {
	first := Observation{Arguments: []string{"format", "input.gooo"}, ExitCode: 0,
		Stdout: "formatted source", Stderr: "", PeakRSSKiB: 11,
		TreeBeforeDigest: "before", TreeAfterDigest: "after", RepositoryWrites: 0}
	replay := first
	replay.PeakRSSKiB = 17
	if !DeterministicReplayEqual(first, replay) || first.PeakRSSKiB != 11 || replay.PeakRSSKiB != 17 {
		t.Fatal("resource observation changed replay identity or was overwritten")
	}
	cases := []struct {
		name   string
		change func(*Observation)
	}{
		{"arguments", func(value *Observation) { value.Arguments = []string{"fix", "input.gooo"} }},
		{"exit-code", func(value *Observation) { value.ExitCode = 1 }},
		{"stdout", func(value *Observation) { value.Stdout = "different source" }},
		{"stderr", func(value *Observation) { value.Stderr = "diagnostic" }},
		{"failure", func(value *Observation) { value.Failure = "INVOCATION_DEADLINE" }},
		{"tree-before", func(value *Observation) { value.TreeBeforeDigest = "other before" }},
		{"tree-after", func(value *Observation) { value.TreeAfterDigest = "other after" }},
		{"writes", func(value *Observation) { value.RepositoryWrites = 1 }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			changed := replay
			test.change(&changed)
			if DeterministicReplayEqual(first, changed) {
				t.Fatal("process contract difference was ignored", test.name)
			}
		})
	}
}
