package toolchainrelease

import (
	"strings"
	"testing"
)

func TestValidateAcceptsExactReport(t *testing.T) {
	report := validReportFixture(t)
	if err := Validate(report, report.HeadSHA); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejectsUnknownDecision(t *testing.T) {
	report := validReportFixture(t)
	report.Decision = "UNRECOGNIZED"
	err := Validate(report, report.HeadSHA)
	if err == nil || !strings.Contains(err.Error(), "TOOLCHAIN_RELEASE_DECISION_UNKNOWN") {
		t.Fatalf("expected fail-closed unknown decision, got %v", err)
	}
}
