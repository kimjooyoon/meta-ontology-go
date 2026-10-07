package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func recordCLIReader(t *testing.T) runSourceReaderWithFiles {
	t.Helper()
	source, err := os.ReadFile("../../examples/language-record-binding/boolean.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile("../../examples/language-record-binding/boolean-input.json")
	if err != nil {
		t.Fatal(err)
	}
	return runSourceReaderWithFiles{"record.gooo": source, "record.json": input}
}

func TestRunSourceRecordInputExecutesDeclaredGraph(t *testing.T) {
	reader := recordCLIReader(t)
	var stdout, stderr bytes.Buffer
	code := runSource([]string{"--json", "--entry", "Capture", "--record-input", "record.json", "record.gooo"}, reader, &stdout, &stderr)
	var report recordPlanReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	if code != exitOK || stderr.Len() != 0 || report.Decision != "PASS" || report.SemanticAdmission != "UNASSESSED" ||
		report.Execution.ApplyCalls != 3 || report.Execution.Deliveries != 2 || report.Execution.Results["Report"].Fields["State"] != "UNKNOWN" ||
		report.Execution.Results["Report"].Fields["Complete"] != false {
		t.Fatalf("code=%d report=%+v stderr=%s", code, report, stderr.String())
	}
}

func TestRunSourceOptionalRecordInputPreservesAbsentAndZeroValues(t *testing.T) {
	source, err := os.ReadFile("../../examples/language-record-binding/optional.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile("../../examples/language-record-binding/optional-input.json")
	if err != nil {
		t.Fatal(err)
	}
	reader := runSourceReaderWithFiles{"optional.gooo": source, "optional.json": input}
	var stdout, stderr bytes.Buffer
	code := runSource([]string{"--json", "--entry", "Capture", "--record-input", "optional.json", "optional.gooo"}, reader, &stdout, &stderr)
	var report recordPlanReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	result := report.Execution.Results["Relay"].Fields
	if code != exitOK || stderr.Len() != 0 || report.Decision != "PASS" ||
		report.Execution.ApplyCalls != 2 || report.Execution.Deliveries != 1 ||
		result["Name"] != "sample" || result["Label"] != "" || result["Complete"] != false || result["Count"] != float64(0) {
		t.Fatalf("code=%d report=%+v stderr=%s", code, report, stderr.String())
	}
	if _, present := result["Note"]; present {
		t.Fatalf("omitted optional field was materialized: %#v", result)
	}
}

func TestRunSourceRecordFailuresNeverClaimAdmission(t *testing.T) {
	for _, entry := range []string{"Review", "Missing"} {
		reader := recordCLIReader(t)
		var stdout, stderr bytes.Buffer
		code := runSource([]string{"--json", "--entry", entry, "--record-input", "record.json", "record.gooo"}, reader, &stdout, &stderr)
		var report recordPlanReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if code != exitFailure || report.Decision != "FAIL_CLOSED" || report.Failure == nil ||
			report.SemanticAdmission != "UNASSESSED" || report.Execution.ApplyCalls != 0 {
			t.Fatalf("entry=%s code=%d report=%+v", entry, code, report)
		}
	}
}

func TestRunSourceRecordModeRejectsMixedInputModes(t *testing.T) {
	if _, err := parseRunSourceArguments([]string{"--entry", "Capture", "--input", "integer.json", "--record-input", "record.json", "record.gooo"}); err == nil {
		t.Fatal("mixed input modes accepted")
	}
}
