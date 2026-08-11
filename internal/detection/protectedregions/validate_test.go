package protectedregions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAcceptsGeneratedSlotsAndHandwrittenRegions(t *testing.T) {
	source := "package fixture\n\n" +
		"//gooo:generated:start id=\"fixture/activity\" kind=\"activity\"\n" +
		"func Activity() int {\n" +
		"\t//gooo:slot:start id=\"fixture/activity/implementation\"\n" +
		"\treturn 1\n" +
		"\t//gooo:slot:end id=\"fixture/activity/implementation\"\n" +
		"}\n" +
		"//gooo:generated:end id=\"fixture/activity\"\n\n" +
		"//gooo:protected:start id=\"fixture/handwritten\"\n" +
		"const Handwritten = 1\n" +
		"//gooo:protected:end id=\"fixture/handwritten\"\n"
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

func TestValidateReportsNestedMarkers(t *testing.T) {
	source := "//gooo:generated:start id=\"outer\"\n" +
		"//gooo:generated:start id=\"inner\"\n" +
		"//gooo:generated:end id=\"inner\"\n" +
		"//gooo:generated:end id=\"outer\"\n"
	report := Validate([]byte(source))
	if !hasIssue(report.Issues, IssueNestedMarker) {
		t.Fatalf("nested generated region was accepted: %#v", report.Issues)
	}

	slotNested := "//gooo:generated:start id=\"activity\"\n" +
		"//gooo:slot:start id=\"outer-slot\"\n" +
		"//gooo:slot:start id=\"inner-slot\"\n" +
		"//gooo:slot:end id=\"inner-slot\"\n" +
		"//gooo:slot:end id=\"outer-slot\"\n" +
		"//gooo:generated:end id=\"activity\"\n"
	if report := Validate([]byte(slotNested)); !hasIssue(report.Issues, IssueNestedMarker) {
		t.Fatalf("nested slot was accepted: %#v", report.Issues)
	}
}

func TestValidateReportsMissingAndDuplicateMarkers(t *testing.T) {
	source := "//gooo:generated:end id=\"orphan\"\n" +
		"//gooo:generated:start id=\"duplicate\"\n" +
		"//gooo:generated:end id=\"duplicate\"\n" +
		"//gooo:generated:start id=\"duplicate\"\n"
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
	report := Validate([]byte("//gooo:slot:start id=\"orphan\"\n//gooo:slot:end id=\"orphan\"\n"))
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
	report := Validate([]byte("//gooo:generated:start id=\"x\"\n"))
	if report.Err() == nil || !strings.Contains(report.Err().Error(), "missing-end") {
		t.Fatalf("missing-end error was not deterministic: %v", report.Err())
	}
	if Check([]byte("//gooo:generated:start id=\"x\"\n")) == nil {
		t.Fatal("Check accepted an unclosed region")
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
