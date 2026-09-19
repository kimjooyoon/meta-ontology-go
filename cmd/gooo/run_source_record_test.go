package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func recordCLIReader(t *testing.T) runSourceReaderWithFiles {
	t.Helper()
	source, err := os.ReadFile("../../examples/language-record-binding/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile("../../examples/language-record-binding/input.json")
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
		report.Execution.ApplyCalls != 3 || report.Execution.Deliveries != 2 || report.Execution.Results["Report"].Fields["State"] != "UNKNOWN" {
		t.Fatalf("code=%d report=%+v stderr=%s", code, report, stderr.String())
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
