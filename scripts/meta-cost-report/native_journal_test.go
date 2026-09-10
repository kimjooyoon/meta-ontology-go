package main

import (
	"bytes"
	"testing"
)

func TestCapturedNativeJournalReconstruction(t *testing.T) {
	first := capturedNativeJournal(t, "first.ndjson")
	replay := capturedNativeJournal(t, "replay.ndjson")
	combined := bytes.Join([][]byte{bytes.TrimSpace(first), bytes.TrimSpace(replay)}, []byte("\n"))
	cases := []struct {
		name      string
		input     []byte
		events    int
		intervals int
		unknowns  int
	}{
		{"completed-native-invocation", first, 24, 12, 0},
		{"interrupted-native-replay", replay, 16, 7, 2},
		{"same-digests-distinct-invocations", combined, 40, 19, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report, err := readCostReport(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if report.Events != tc.events || len(report.Rows) != tc.intervals ||
				report.UnpairedStarts != tc.unknowns || len(report.Unknowns) != tc.unknowns ||
				report.UnmeasuredEvents != 0 || report.UnknownReturns != 0 {
				t.Fatalf("incorrect native reconstruction: %+v", report)
			}
			if report.Authenticity != "UNVERIFIED" || report.Improvement != "UNKNOWN" ||
				report.Scope != "DIAGNOSTIC_INTERVALS_ONLY_NO_ADDITIVE_TOTAL" {
				t.Fatalf("diagnostic input gained authority: %+v", report)
			}
			assertCapturedBoundaryUnknowns(t, report.Unknowns)
		})
	}
}

func assertCapturedBoundaryUnknowns(t *testing.T, unknowns []boundaryUnknown) {
	t.Helper()
	starts := []uint64{11, 16}
	steps := []string{"ACTION_ENTERED", "PROCESS_CALL_ENTERED"}
	kinds := []string{"action", "verifier"}
	passes := []string{"selected", "first"}
	for i, unknown := range unknowns {
		if unknown.State != "UNKNOWN" || unknown.Stage != "DRIVER_BOUNDARY" || unknown.Step != steps[i] ||
			unknown.Reason != "MATCHING_TERMINAL_OBSERVATION_MISSING" || unknown.UnknownClass != "DIRECT_MISSING" ||
			unknown.NextOperation != "OBSERVE_MATCHING_TERMINAL_EVENT" || unknown.BlockedBy == nil || len(unknown.BlockedBy) != 0 {
			t.Fatalf("missing terminal gained an inferred cause: %+v", unknown)
		}
		if unknown.Invocation != capturedReplayInvocation || unknown.Head != capturedJournalHead ||
			unknown.StartEvent != starts[i] || unknown.Kind != kinds[i] || unknown.Pass != passes[i] ||
			unknown.Activity != "CollapseAssignReturn" || unknown.Operation != 2 ||
			unknown.Subject != "scripts/meta-execution/collapsefixture/collapse_fixture.go:3:CollapseFixture" ||
			unknown.Indicator != "sha256:cebe8569062bdb2febda5536e88f05bb6451ede7dc48f6e99f9224cd75d4d18b" ||
			unknown.Plan != "c4f3516002b5cd8a6b2cce1cf5ad8f1b04ec74ee29d0ad8df5059341f858e235" ||
			unknown.Manifest != "a628347661f1635cf97e8ef753970fd037f1ac6bf0c217604e540c0bc663b9b2" ||
			unknown.Source != "b886df4511dab6fcc9f37a9a661594f0fd8bb24f43f92337bb30146354177a0a" ||
			unknown.Semantic != "b9a780ed3ba67f58bd91a86e83280ef8b98fd3b38893dc11ddd69169c85e012e" {
			t.Fatalf("UNKNOWN lost its native invocation/activity binding: %+v", unknown)
		}
	}
}
