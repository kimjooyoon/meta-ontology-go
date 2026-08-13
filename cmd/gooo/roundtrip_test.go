package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRunRoundTripReportsBXEvidence(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runRoundTrip([]string{"--json", "fixture.gooo"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("roundtrip = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	var report jsonReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ok" || report.Equivalent == nil || !*report.Equivalent || report.GetPut == nil || !*report.GetPut || report.PutGet == nil || !*report.PutGet || report.OriginalSemanticHash != report.RoundTrippedSemanticHash {
		t.Fatalf("roundtrip report lacks equivalent semantic evidence: %#v", report)
	}
}

func TestRunRoundTripDoesNotWriteSource(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runRoundTrip([]string{"fixture.gooo"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("read-only roundtrip = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}
