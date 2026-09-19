package main

import (
	"strings"
	"testing"
)

func TestMissingReturnHasBoundSixFieldUnknown(t *testing.T) {
	report, err := readCostReport(strings.NewReader(startEvent))
	if err != nil || len(report.Unknowns) != 1 {
		t.Fatalf("missing bound unknown: %+v %v", report, err)
	}
	got := report.Unknowns[0]
	if got.State != "UNKNOWN" || got.Stage != "DRIVER_BOUNDARY" ||
		got.Step != "PROCESS_CALL_ENTERED" || got.Reason != "MATCHING_TERMINAL_OBSERVATION_MISSING" ||
		got.UnknownClass != "DIRECT_MISSING" || got.NextOperation != "OBSERVE_MATCHING_TERMINAL_EVENT" ||
		got.BlockedBy == nil || len(got.BlockedBy) != 0 || got.Invocation != "one" ||
		got.Activity != "gooo://activity" || got.StartEvent != 1 {
		t.Fatalf("invented or unbound cause: %+v", got)
	}
}

func TestMatchedReturnDoesNotInventUnknown(t *testing.T) {
	report, err := readCostReport(strings.NewReader(startEvent + "\n" + returnEvent))
	if err != nil || report.Unknowns == nil || len(report.Unknowns) != 0 {
		t.Fatalf("matched interval gained unknown: %+v %v", report, err)
	}
}
