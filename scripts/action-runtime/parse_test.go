package main

import "testing"

func TestParseWorkflowPreservesActionInputs(t *testing.T) {
	source := []byte(`jobs:
  check:
    steps:
      - uses: actions/checkout@v5
      - uses: local/action@v1
      - uses: actions/download-artifact@v7
        with:
          run-id: 42
          name: evidence
          if-no-files-found: error
      - run: true
`)
	sites := parseWorkflow(source)
	if len(sites) != 2 {
		t.Fatalf("sites = %d, want 2", len(sites))
	}
	if sites[0].Action != "actions/checkout" || sites[0].Line != 4 {
		t.Fatalf("first site = %#v", sites[0])
	}
	got := sites[1].Inputs
	want := []string{"if-no-files-found", "name", "run-id"}
	if len(got) != len(want) {
		t.Fatalf("inputs = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("inputs = %v, want %v", got, want)
		}
	}
}

func TestParseMajorRejectsUncataloguedReferences(t *testing.T) {
	if major, ok := parseMajor("v8.1.0"); !ok || major != 8 {
		t.Fatalf("parseMajor(v8.1.0) = %d, %v", major, ok)
	}
	if _, ok := parseMajor("main"); ok {
		t.Fatal("parseMajor(main) unexpectedly succeeded")
	}
}
