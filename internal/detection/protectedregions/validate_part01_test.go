package protectedregions

import (
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
