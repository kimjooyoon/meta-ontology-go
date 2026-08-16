package protectedregions

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestValidateAcceptsGeneratedSlotsAndHandwrittenRegions(t *testing.T) {
	source := "package fixture\n\n" +
		"//gooo:generated:start id=\"fixture://activity\" kind=\"activity\"\n" +
		"func Activity() int {\n" +
		"\t//gooo:slot:start id=\"fixture://activity/implementation\"\n" +
		"\treturn 1\n" +
		"\t//gooo:slot:end id=\"fixture://activity/implementation\"\n" +
		"}\n" +
		"//gooo:generated:end id=\"fixture://activity\" kind=\"activity\"\n\n" +
		"//gooo:protected:start id=\"fixture://handwritten\"\n" +
		"const Handwritten = 1\n" +
		"//gooo:protected:end id=\"fixture://handwritten\"\n"
	report := Validate([]byte(source))
	if !report.Valid() {
		t.Fatalf("valid markers rejected: %v", report.Err())
	}
	if len(report.Regions) != 3 {
		t.Fatalf("expected generated, slot, and protected regions, got %#v", report.Regions)
	}
	if report.Regions[0].Kind != Generated || report.Regions[1].Kind != Slot || report.Regions[2].Kind != Handwritten {
		t.Fatalf("regions are not in source order: %#v", report.Regions)
	}
	if got := string(report.Regions[1].Body([]byte(source))); got != "\treturn 1\n" {
		t.Fatalf("slot body was not recorded: %q", got)
	}
}

func TestValidateAcceptsCurrentLegacyGeneratorMarkers(t *testing.T) {
	source := "//gooo:generated begin fixture://legacy\nbody\n//gooo:generated end\n"
	report := Validate([]byte(source))
	if !report.Valid() || len(report.Regions) != 1 || report.Regions[0].ID != "fixture://legacy" {
		t.Fatalf("legacy markers were not accepted: %#v", report)
	}
}

func TestValidateReportsNestedMarkers(t *testing.T) {
	source := "//gooo:generated:start id=\"fixture://outer\" kind=\"activity\"\n" +
		"//gooo:generated:start id=\"fixture://inner\" kind=\"activity\"\n" +
		"//gooo:generated:end id=\"fixture://inner\" kind=\"activity\"\n" +
		"//gooo:generated:end id=\"fixture://outer\" kind=\"activity\"\n"
	report := Validate([]byte(source))
	if !hasIssue(report.Issues, IssueNestedMarker) {
		t.Fatalf("nested generated region was accepted: %#v", report.Issues)
	}

	slotNested := "//gooo:generated:start id=\"fixture://activity\" kind=\"activity\"\n" +
		"//gooo:slot:start id=\"fixture://outer-slot\"\n" +
		"//gooo:slot:start id=\"fixture://inner-slot\"\n" +
		"//gooo:slot:end id=\"fixture://inner-slot\"\n" +
		"//gooo:slot:end id=\"fixture://outer-slot\"\n" +
		"//gooo:generated:end id=\"fixture://activity\" kind=\"activity\"\n"
	if report := Validate([]byte(slotNested)); !hasIssue(report.Issues, IssueNestedMarker) {
		t.Fatalf("nested slot was accepted: %#v", report.Issues)
	}
}

func TestValidateReportsMissingAndDuplicateMarkers(t *testing.T) {
	source := "//gooo:generated:end id=\"fixture://orphan\" kind=\"activity\"\n" +
		"//gooo:generated:start id=\"fixture://duplicate\" kind=\"activity\"\n" +
		"//gooo:generated:end id=\"fixture://duplicate\" kind=\"activity\"\n" +
		"//gooo:generated:start id=\"fixture://duplicate\" kind=\"activity\"\n"
	report := Validate([]byte(source))
	if !hasIssue(report.Issues, IssueMissingStart) || !hasIssue(report.Issues, IssueMissingEnd) {
		t.Fatalf("missing marker was not reported: %#v", report.Issues)
	}
	if !hasIssue(report.Issues, IssueDuplicateMarker) {
		t.Fatalf("duplicate marker was not reported: %#v", report.Issues)
	}
	missingID := Validate([]byte("//gooo:generated:start\n//gooo:generated:end\n"))
	if !hasIssue(missingID.Issues, IssueMissingID) {
		t.Fatalf("marker without stable ID was accepted: %#v", missingID.Issues)
	}
}

func TestValidateReportsSlotsOutsideGeneratedRegion(t *testing.T) {
	report := Validate([]byte("//gooo:slot:start id=\"fixture://orphan\"\n//gooo:slot:end id=\"fixture://orphan\"\n"))
	if !hasIssue(report.Issues, IssueSlotOutsideGenerated) {
		t.Fatalf("orphan slot was accepted: %#v", report.Issues)
	}
}

func TestLocalityFixtureAllowsGeneratedChange(t *testing.T) {
	before := readFixture(t, "before.go")
	after := readFixture(t, "after.go")
	report := ValidateLocality(before, after)
	if !report.Valid() {
		t.Fatalf("local generated change was rejected: %v", report.Err())
	}
}

func TestLocalityRejectsProtectedOverwrite(t *testing.T) {
	before := readFixture(t, "before.go")
	after := strings.Replace(string(before), "return 7", "return 8", 1)
	report := ValidateLocality(before, []byte(after))
	if report.Valid() || !hasLocalityIssue(report.Violations, LocalityProtectedChange) {
		t.Fatalf("slot overwrite was accepted: %#v", report.Violations)
	}

	after = strings.Replace(string(before), "var Keep = 7", "var Keep = 8", 1)
	report = ValidateLocality(before, []byte(after))
	if report.Valid() || !hasLocalityIssue(report.Violations, LocalityUnownedChange) {
		t.Fatalf("unmarked handwritten overwrite was accepted: %#v", report.Violations)
	}
}

func TestCheckReturnsDeterministicErrors(t *testing.T) {
	report := Validate([]byte("//gooo:generated:start id=\"fixture://x\" kind=\"activity\"\n"))
	if report.Err() == nil || !strings.Contains(report.Err().Error(), "missing-end") {
		t.Fatalf("missing-end error was not deterministic: %v", report.Err())
	}
	if Check([]byte("//gooo:generated:start id=\"fixture://x\" kind=\"activity\"\n")) == nil {
		t.Fatal("Check accepted an unclosed region")
	}
}

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

func TestLocalityViolationOrderAndNoMutation(t *testing.T) {
	before := []byte("package fixture\n" +
		"//gooo:protected:start id=\"fixture://z\"\n" +
		"const Z = 1\n" +
		"//gooo:protected:end id=\"fixture://z\"\n" +
		"//gooo:protected:start id=\"fixture://a\"\n" +
		"const A = 1\n" +
		"//gooo:protected:end id=\"fixture://a\"\n" +
		"//gooo:generated:start id=\"fixture://activity\" kind=\"activity\"\n" +
		"func Activity() int { return 1 }\n" +
		"//gooo:generated:end id=\"fixture://activity\" kind=\"activity\"\n")
	after := []byte(strings.Replace(strings.Replace(string(before), "const Z = 1", "const Z = 2", 1), "const A = 1", "const A = 2", 1))
	beforeCopy := append([]byte(nil), before...)
	afterCopy := append([]byte(nil), after...)
	report := ValidateLocality(before, after)
	if report.Valid() || len(report.Violations) != 2 {
		t.Fatalf("violations = %#v", report.Violations)
	}
	if report.Violations[0].ID != "fixture://a" || report.Violations[1].ID != "fixture://z" {
		t.Fatalf("violations were not sorted by stable ID: %#v", report.Violations)
	}
	if !bytes.Equal(before, beforeCopy) || !bytes.Equal(after, afterCopy) {
		t.Fatal("locality validation mutated its inputs")
	}
}

func TestLocalityRejectsGeneratedMarkerLineChanges(t *testing.T) {
	before := readFixture(t, "before.go")
	after := strings.Replace(string(before), "//gooo:generated:start id=\"fixture://activity\"", " //gooo:generated:start id=\"fixture://activity\"", 1)
	report := ValidateLocality(before, []byte(after))
	if !report.Before.Valid() || !report.After.Valid() || !hasLocalityIssue(report.Violations, LocalityUnownedChange) {
		t.Fatalf("generated marker line change was accepted without a locality violation: %#v", report)
	}
}

func hasIssue(issues []Issue, want IssueKind) bool {
	for _, issue := range issues {
		if issue.Kind == want {
			return true
		}
	}
	return false
}

func hasLocalityIssue(issues []LocalityIssue, want LocalityIssueKind) bool {
	for _, issue := range issues {
		if issue.Kind == want {
			return true
		}
	}
	return false
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", "locality", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return source
}
