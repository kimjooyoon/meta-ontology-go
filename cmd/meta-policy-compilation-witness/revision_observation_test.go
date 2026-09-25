package main

import (
	"flag"
	"io"
	"testing"
)

func TestRevisionObservationModeIsExplicitAndDisjoint(t *testing.T) {
	cases := []struct {
		name string
		args []string
		mode bool
		fail bool
	}{
		{"legacy", []string{"-cases", "cases.json"}, false, false},
		{"revision", []string{"-policy", "policy.gooo", "-observe-revision", "request.json"}, true, false},
		{"profile", []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "-profile-package", "custom", "-profile-namespace", "custom"}, true, false},
		{"mixed-cases", []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "-cases", "cases.json"}, true, true},
		{"mixed-output", []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "-output", "out"}, true, true},
		{"mixed-project-root", []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "-profile-project-root", "."}, true, true},
		{"positional", []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "extra"}, true, true},
		{"empty-request", []string{"-policy", "policy.gooo", "-observe-revision", ""}, true, true},
		{"missing-policy", []string{"-observe-revision", "request.json"}, true, true},
	}
	for _, current := range cases {
		t.Run(current.name, func(t *testing.T) {
			flags := flag.NewFlagSet(current.name, flag.ContinueOnError)
			flags.SetOutput(io.Discard)
			for _, name := range []string{"observe-revision", "policy", "cases", "output", "profile-package", "profile-namespace", "profile-project-root"} {
				flags.String(name, "", "")
			}
			if err := flags.Parse(current.args); err != nil {
				t.Fatal(err)
			}
			mode, err := revisionObservationMode(flags)
			if mode != current.mode || (err != nil) != current.fail {
				t.Fatalf("mode boundary differs: mode=%v error=%v", mode, err)
			}
		})
	}
}
