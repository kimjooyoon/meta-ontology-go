package main

import "testing"

func TestSourceControlInputOptions(t *testing.T) {
	base := []string{"--activity", "Adjust", "--inputs-for", "source/control:Inspect"}
	if options, err := parseBodyOriginOptions(append(append([]string{}, base...), "--inputs-shape", "source-control-columns", "source.gooo")); err != nil || options.flags["--inputs-shape"] != "source-control-columns" {
		t.Fatal("explicit control projection", options, err)
	}
	for _, args := range [][]string{
		{"--activity", "Adjust", "--inputs-shape", "source-control-columns", "source.gooo"},
		append(append([]string{}, base...), "--inputs-shape", "unknown", "source.gooo"),
		append(append([]string{}, base...), "--inputs-shape", "source-control-columns", "--field", "urn:field", "source.gooo"),
		append(append([]string{}, base...), "--inputs-shape", "source-control-columns", "--fill-candidate", "candidate", "source.gooo"),
	} {
		if _, err := parseBodyOriginOptions(args); err == nil {
			t.Fatal("incompatible projection options accepted", args)
		}
	}
}
