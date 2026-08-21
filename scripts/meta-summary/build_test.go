package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBuildIsDeterministicAndBounded(t *testing.T) {
	opts := fixtureOptions(t, "BOUND")
	first, firstReport, err := build(opts)
	if err != nil {
		t.Fatal(err)
	}
	second, secondReport, err := build(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) || !reflect.DeepEqual(firstReport, secondReport) {
		t.Fatal("summary projection is not deterministic")
	}
	if firstReport.SchemaVersion != summarySchema || firstReport.Decision != "PASS" {
		t.Fatalf("unexpected report tuple: %#v", firstReport)
	}
	if firstReport.OutputBytes != len(first) || firstReport.OutputBytes > opts.LimitBytes {
		t.Fatalf("invalid byte metric: report=%d actual=%d", firstReport.OutputBytes, len(first))
	}
	if len(firstReport.Indicators) != 4 || len(firstReport.Artifacts) != 5 {
		t.Fatalf("incomplete indicator projection: %#v", firstReport)
	}
}

func TestBuildRejectsUnboundProvenance(t *testing.T) {
	opts := fixtureOptions(t, "BROKEN")
	if _, _, err := build(opts); err == nil {
		t.Fatal("unbound provenance was accepted")
	}
}

func TestBuildRejectsExceededBudget(t *testing.T) {
	opts := fixtureOptions(t, "BOUND")
	opts.LimitBytes = 1
	if _, report, err := build(opts); err == nil || report.Decision != "FAIL" {
		t.Fatalf("exceeded summary budget was not rejected: report=%#v err=%v", report, err)
	}
}
