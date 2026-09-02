package main

import "testing"

func exactReports() (densityReport, extractionReport, splitReport) {
	return densityReport{Schema: densitySchema, SourceSHA: "abc", Subjects: []densitySubject{
		{Logical: "a.go", Status: "applied"}, {Logical: "blocked.go", Status: "blocked"},
	}}, extractionReport{Schema: extractionSchema, SourceSHA: "abc", Subjects: []extractionSubject{
		{Files: []string{"a.go", "c.go"}},
	}}, splitReport{Schema: splitSchema, SourceSHA: "abc", Decision: "FIXED_POINT", Resolution: "EXACT", Exact: true}
}

func TestReduceUnionsDensityAndExtractionReceipts(t *testing.T) {
	density, extraction, split := exactReports()
	report := reduce("abc", density, extraction, split, []string{"c.go", "a.go"}, nil)
	if report.Decision != "PASS" || report.Resolution != "EXACT" || !report.Exact {
		t.Fatalf("decision=%s resolution=%s reason=%s", report.Decision, report.Resolution, report.Reason)
	}
	want := []string{"a.go", "c.go"}
	if !equalPaths(report.Expected, want) || !equalPaths(report.Observed, want) {
		t.Fatalf("expected=%v observed=%v", report.Expected, report.Observed)
	}
	c := report.Coordinates
	if c.SourceReceipts != 3 || c.DensityPaths != 1 || c.ExtractionPaths != 2 ||
		c.OverlapPaths != 1 || c.ExpectedPaths != 2 || c.Unknowns != 0 {
		t.Fatalf("coordinates=%#v", c)
	}
}

func TestReduceKnownMismatchKeepsExactResolution(t *testing.T) {
	density, extraction, split := exactReports()
	report := reduce("abc", density, extraction, split, []string{"a.go"}, nil)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "EXACT" ||
		report.Reason != "WRITE_SET_NOT_EXACT" || report.Coordinates.Unknowns != 0 {
		t.Fatalf("report=%#v", report)
	}
}

func TestReduceRejectsUntrackedPath(t *testing.T) {
	density, extraction, split := exactReports()
	report := reduce("abc", density, extraction, split, []string{"a.go", "c.go", "rogue.go"}, []string{"rogue.go"})
	if report.Reason != "WRITE_SET_NOT_EXACT" || report.Coordinates.UnclassifiedPaths != 1 {
		t.Fatalf("report=%#v", report)
	}
}

func TestReduceAcceptsSplitDeclaredCreation(t *testing.T) {
	density, extraction, split := exactReports()
	split.Decision = "PASS"
	split.Subjects = []extractionSubject{{Files: []string{"d.go", "d_split02.go"}, Created: []string{"d_split02.go"}}}
	split.Coordinates = splitCoordinates{Selected: 1, Applied: 1, Changed: 2, Created: 1}
	report := reduce("abc", density, extraction, split,
		[]string{"a.go", "c.go", "d.go", "d_split02.go"}, []string{"d_split02.go"})
	if report.Decision != "PASS" || report.Coordinates.CreatedPaths != 1 || report.Coordinates.SplitPaths != 2 ||
		report.Coordinates.UnclassifiedPaths != 0 {
		t.Fatalf("report=%#v", report)
	}
}
