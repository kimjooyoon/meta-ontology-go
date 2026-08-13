package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRunQueryReturnsStableSemanticProjection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runQuery([]string{"--json", "fixture.gooo", "--kind", "activity", "--predicate", "prov:used"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("query = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	var report jsonReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ok" || len(report.Nodes) != 1 || report.Nodes[0].Kind != "Activity" || len(report.Facts) != 1 {
		t.Fatalf("unexpected query projection: %#v", report)
	}
	if report.Facts[0].Predicate != "used" {
		t.Fatalf("predicate filter was not applied: %#v", report.Facts)
	}
}

func TestRunQueryUnknownIDHasStableFailureCode(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runQuery([]string{"--json", "fixture.gooo", "--id", "billing://entity/missing"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitFailure || stderr.Len() != 0 {
		t.Fatalf("unknown query ID = %d, stderr=%q", code, stderr.String())
	}
	var report jsonReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "error" || len(report.Diagnostics) != 1 || report.Diagnostics[0].Code != "query.invalid" {
		t.Fatalf("unexpected unknown-ID report: %#v", report)
	}
}
