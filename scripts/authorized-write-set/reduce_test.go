package main

import "testing"

func exactReports() (densityReport, extractionReport) {
	return densityReport{Schema: densitySchema, SourceSHA: "abc", Subjects: []densitySubject{
		{Logical: "a.go", Status: "applied"}, {Logical: "blocked.go", Status: "blocked"},
	}}, extractionReport{Schema: extractionSchema, SourceSHA: "abc", Subjects: []extractionSubject{
		{Files: []string{"a.go", "c.go"}},
	}}
}

func TestReduceUnionsDensityAndExtractionReceipts(t *testing.T) {
	density, extraction := exactReports()
	report := reduce("abc", density, extraction, []string{"c.go", "a.go"}, 0)
	if report.Decision != "PASS" || report.Resolution != "EXACT" || !report.Exact {
		t.Fatalf("decision=%s resolution=%s reason=%s", report.Decision, report.Resolution, report.Reason)
	}
	want := []string{"a.go", "c.go"}
	if !equalPaths(report.Expected, want) || !equalPaths(report.Observed, want) {
		t.Fatalf("expected=%v observed=%v", report.Expected, report.Observed)
	}
	c := report.Coordinates
	if c.SourceReceipts != 2 || c.DensityPaths != 1 || c.ExtractionPaths != 2 ||
		c.OverlapPaths != 1 || c.ExpectedPaths != 2 || c.Unknowns != 0 {
		t.Fatalf("coordinates=%#v", c)
	}
}

func TestReduceKnownMismatchKeepsExactResolution(t *testing.T) {
	density, extraction := exactReports()
	report := reduce("abc", density, extraction, []string{"a.go"}, 0)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "EXACT" ||
		report.Reason != "WRITE_SET_NOT_EXACT" || report.Coordinates.Unknowns != 0 {
		t.Fatalf("report=%#v", report)
	}
}

func TestReduceRejectsUntrackedPath(t *testing.T) {
	density, extraction := exactReports()
	report := reduce("abc", density, extraction, []string{"a.go", "c.go"}, 1)
	if report.Reason != "WRITE_SET_NOT_EXACT" || report.Coordinates.UntrackedPaths != 1 {
		t.Fatalf("report=%#v", report)
	}
}
