package main

import (
	"strings"
	"testing"
)

const startEvent = `{"schema":"gooo/meta-execution-driver-boundary/v1","invocation_id":"one","event_sequence":1,"activity":"gooo://activity","pass":"first","boundary":"PROCESS_CALL_ENTERED","cost":{"state":"STARTED"}}`
const returnEvent = `{"schema":"gooo/meta-execution-driver-boundary/v1","invocation_id":"one","event_sequence":2,"activity":"gooo://activity","pass":"first","boundary":"PROCESS_RETURNED","cost":{"state":"OBSERVED","started_at_event":1,"elapsed_ns":123}}`

func TestReportPreservesBoundIntervalWithoutImprovementClaim(t *testing.T) {
	report, err := readCostReport(strings.NewReader(startEvent + "\n" + returnEvent))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Rows) != 1 || report.Rows[0].Elapsed != 123 || report.UnpairedStarts != 0 ||
		report.Improvement != "UNKNOWN" || report.Authenticity != "UNVERIFIED" {
		t.Fatalf("incorrect report: %+v", report)
	}
}

func TestReportRejectsInvalidBindingsAndDuplicateEvents(t *testing.T) {
	cases := []string{
		startEvent + "\n" + startEvent,
		returnEvent,
		startEvent + "\n" + strings.Replace(returnEvent, `"first"`, `"replay"`, 1),
		startEvent + "\n" + strings.Replace(returnEvent, `"gooo://activity"`, `"other"`, 1),
		startEvent + "\n" + strings.Replace(returnEvent, `123`, `-1`, 1),
		startEvent + "\n" + strings.Replace(returnEvent, `"OBSERVED"`, `"FIXED_POINT"`, 1),
	}
	for index, input := range cases {
		if _, err := readCostReport(strings.NewReader(input)); err == nil {
			t.Errorf("accepted invalid case %d", index)
		}
	}
}

func TestReportRetainsUnfinishedAndLegacyObservations(t *testing.T) {
	report, err := readCostReport(strings.NewReader(startEvent))
	if err != nil || report.UnpairedStarts != 1 || len(report.Rows) != 0 {
		t.Fatalf("unfinished interval: %+v, %v", report, err)
	}
	legacy := strings.Replace(startEvent, `,"cost":{"state":"STARTED"}`, "", 1)
	report, err = readCostReport(strings.NewReader(legacy))
	if err != nil || report.UnmeasuredEvents != 1 || len(report.Rows) != 0 {
		t.Fatalf("legacy event: %+v, %v", report, err)
	}
}
