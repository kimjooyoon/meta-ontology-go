package protectedregions

import (
	"strings"
	"testing"
)

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
