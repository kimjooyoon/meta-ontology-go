package main

import "testing"

func TestInputSignatureExportOptions(t *testing.T) {
	base := []string{"--activity", "Difference", "--inputs-for", "source/control:InspectInputs", "--inputs-shape", "input-declaration-columns"}
	if _, err := parseBodyOriginOptions(append(append([]string{}, base...), "source.gooo")); err != nil {
		t.Fatal(err)
	}
	for _, extra := range [][]string{
		{"--field", "urn:field"}, {"--include-control"}, {"--fill-candidate", "candidate"},
		{"--fill-candidates"}, {"--format", "text"},
	} {
		args := append(append(append([]string{}, base...), extra...), "source.gooo")
		if _, err := parseBodyOriginOptions(args); err == nil {
			t.Fatal("incompatible signature options accepted", args)
		}
	}
	if _, err := parseBodyOriginOptions([]string{"--activity", "Difference", "--inputs-shape", "input-declaration-columns", "source.gooo"}); err == nil {
		t.Fatal("missing consumer accepted")
	}
}
