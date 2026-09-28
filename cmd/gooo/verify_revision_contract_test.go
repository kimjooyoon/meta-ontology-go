package main

import "testing"

func TestParseVerifyRevisionContractArgumentsAcceptsExpectedOutputs(t *testing.T) {
	options, err := parseVerifyRevisionContractArguments([]string{
		"baseline.gooo",
		"candidate.gooo",
		"--revision", "revision.json",
		"--evaluation", "evaluation.json",
		"--activity", "Observe",
		"--inputs", "inputs.json",
		"--expected-outputs", "expected.json",
		"--out", "out",
	})
	if err != nil {
		t.Fatal(err)
	}
	if options.expectedOutputs != "expected.json" {
		t.Fatalf("expected outputs path = %q", options.expectedOutputs)
	}
}
