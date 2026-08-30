package toolchaincli

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func Validate(report Report, expectedHead string) error {
	if report.Schema != ReportSchema || report.Source.ExpectedHeadSHA != expectedHead ||
		report.ReportDigest == "" || report.ReportDigest != reportDigest(report) || !validHead(expectedHead) {
		return fmt.Errorf("toolchain CLI report binding invalid")
	}
	if report.ResourceObservationMode != "RUNNER_SCOPED_NONDETERMINISTIC" ||
		report.ResourceMeasurementReplayAuthority || report.PerformanceImprovement != "UNKNOWN" {
		return fmt.Errorf("toolchain CLI resource boundary invalid")
	}
	if report.Decision != DecisionPass && report.Decision != DecisionClosed {
		return fmt.Errorf("toolchain CLI decision unknown")
	}
	if report.Resolution != ResolutionExact && report.Resolution != ResolutionLower {
		return fmt.Errorf("toolchain CLI resolution unknown")
	}
	if len(report.Cases) != FixedTotal || len(report.Indicators) != FixedIndicators ||
		len(report.Proofs) != 3 || len(report.Stages) != 5 || report.MutationAuthorized {
		return fmt.Errorf("toolchain CLI report shape invalid")
	}
	for _, result := range report.Cases {
		if result.EvidenceDigest != caseDigest(result) {
			return fmt.Errorf("toolchain CLI case digest invalid")
		}
	}
	if report.Decision == DecisionPass && (report.Resolution != ResolutionExact ||
		report.Summary.Satisfied != FixedTotal || report.Summary.Unresolved != 0 ||
		report.RepositoryWrites != 0 || report.Summary.ResourceObservations != ExpectedRuns ||
		report.Summary.PeakRSSKiB <= 0 || !allIndicators(report.Indicators) || !allProofs(report.Proofs)) {
		return fmt.Errorf("toolchain CLI exact evidence incomplete")
	}
	if report.Resolution == ResolutionLower && (report.Decision != DecisionClosed || report.Summary.Unresolved == 0) {
		return fmt.Errorf("toolchain CLI lower resolution invalid")
	}
	return nil
}

func reportDigest(report Report) string {
	report.ReportDigest = ""
	return digestJSON(report)
}

func validHead(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
