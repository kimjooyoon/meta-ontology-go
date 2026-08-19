package protectedregions

import (
	"reflect"
	"testing"
)

func TestValidateRejectsMalformedCurrentMarkers(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   IssueKind
	}{
		{name: "line boundary", source: "//gooo:generated:startling id=\"fixture://x\" kind=\"activity\"\n", want: IssueInvalidMarker},
		{name: "unknown attribute", source: "//gooo:slot:start id=\"fixture://x\" owner=\"user\"\n", want: IssueInvalidMarker},
		{name: "missing generated kind", source: "//gooo:generated:start id=\"fixture://x\"\n", want: IssueMissingKind},
		{name: "invalid generated kind", source: "//gooo:generated:start id=\"fixture://x\" kind=\"agent\"\n", want: IssueInvalidMarker},
		{name: "legacy end attributes", source: "//gooo:generated begin fixture://x\n//gooo:generated end fixture://x\n", want: IssueInvalidMarker},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if report := Validate([]byte(test.source)); !hasIssue(report.Issues, test.want) {
				t.Fatalf("issues = %#v, want %s", report.Issues, test.want)
			}
		})
	}
}
func TestValidateReportsStableIDsAcrossMarkerKindsAsDuplicates(t *testing.T) {
	source := "//gooo:generated:start id=\"fixture://same\" kind=\"activity\"\n" +
		"//gooo:slot:start id=\"fixture://same\"\n" +
		"//gooo:slot:end id=\"fixture://same\"\n" +
		"//gooo:generated:end id=\"fixture://same\" kind=\"activity\"\n"
	if report := Validate([]byte(source)); !hasIssue(report.Issues, IssueDuplicateMarker) {
		t.Fatalf("stable ID collision was accepted: %#v", report.Issues)
	}
}
func TestValidateIsDeterministicAndMissingEndsFollowSourceOrder(t *testing.T) {
	source := "//gooo:generated:start id=\"fixture://first\" kind=\"activity\"\n" +
		"//gooo:generated:start id=\"fixture://second\" kind=\"activity\"\n"
	wantLines := []int{1, 2}
	var first Report
	for run := 0; run < 20; run++ {
		report := Validate([]byte(source))
		if run == 0 {
			first = report
		} else if !reflect.DeepEqual(first, report) {
			t.Fatalf("run %d changed report: first=%#v got=%#v", run, first, report)
		}
		missingLines := make([]int, 0, len(wantLines))
		for _, issue := range report.Issues {
			if issue.Kind == IssueMissingEnd {
				missingLines = append(missingLines, issue.Line)
			}
		}
		if !reflect.DeepEqual(missingLines, wantLines) {
			t.Fatalf("issues = %#v, want missing ends at lines %v", report.Issues, wantLines)
		}
	}
}
