package main

import (
	"flag"
	"io"
	"testing"
)

func TestGoooRevisionOperationFlagRequiresExplicitObservation(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		valid bool
	}{
		{"bound", []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "-revision-operation", "operation.gooo"}, true},
		{"no-request", []string{"-policy", "policy.gooo", "-revision-operation", "operation.gooo"}, false},
		{"empty-operation", []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "-revision-operation", ""}, false},
		{"no-policy", []string{"-observe-revision", "request.json", "-revision-operation", "operation.gooo"}, false},
		{"mixed-mode", []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "-revision-operation", "operation.gooo", "-generate"}, false},
		{"positional", []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "-revision-operation", "operation.gooo", "extra"}, false},
	}
	for _, candidate := range cases {
		t.Run(candidate.name, func(t *testing.T) {
			flags := flag.NewFlagSet("bound-revision", flag.ContinueOnError)
			flags.SetOutput(io.Discard)
			for _, name := range []string{"policy", "observe-revision", "revision-operation", "profile-package", "profile-namespace"} {
				flags.String(name, "", "")
			}
			flags.Bool("generate", false, "")
			if err := flags.Parse(candidate.args); err != nil {
				t.Fatal(err)
			}
			requested, err := revisionObservationMode(flags)
			if !requested || (err == nil) != candidate.valid {
				t.Fatalf("requested=%v error=%v valid=%v", requested, err, candidate.valid)
			}
		})
	}
}
