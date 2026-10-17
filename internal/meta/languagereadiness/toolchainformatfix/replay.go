package toolchainformatfix

import (
	"fmt"
	"reflect"
)

const caseProofOperation = "reject-unknown-and-write-boundaries"

// CompareReplayReports checks report equality after validating both raw records.
// Equality covers retained observations and origins; it grants no execution or
// reuse authority and can describe two equally unresolved reports.
func CompareReplayReports(first, replay Report, expectedHead string) error {
	if err := validateReplayReport(first, expectedHead); err != nil {
		return fmt.Errorf("first format/fix report: %w", err)
	}
	if err := validateReplayReport(replay, expectedHead); err != nil {
		return fmt.Errorf("replay format/fix report: %w", err)
	}
	if !reflect.DeepEqual(reportReplayView(first), reportReplayView(replay)) {
		return fmt.Errorf("FAIL_CLOSED: toolchain format/fix replay mismatch")
	}
	return nil
}

func reportReplayView(report Report) Report {
	report.ReportDigest = ""
	report.Cases = append([]CaseResult(nil), report.Cases...)
	for index := range report.Cases {
		// These digests bind the raw RSS values and were checked before projection.
		report.Cases[index].EvidenceDigest = ""
		report.Cases[index].First.PeakRSSKiB = 0
		report.Cases[index].Replay.PeakRSSKiB = 0
	}
	report.Proofs = append([]Proof(nil), report.Proofs...)
	for index := range report.Proofs {
		if report.Proofs[index].Choice == "REGRESSION" {
			report.Proofs[index].EvidenceDigest = ""
		}
	}
	return report
}

func validateReplayReport(report Report, expectedHead string) error {
	if err := Validate(report, expectedHead); err != nil {
		return err
	}
	caseProofs := 0
	for _, proof := range report.Proofs {
		if proof.Choice != "REGRESSION" {
			continue
		}
		caseProofs++
		if proof.MetaOperation != caseProofOperation || proof.EvidenceDigest != digestJSON(report.Cases) {
			return fmt.Errorf("format/fix case proof binding invalid")
		}
	}
	if caseProofs != 1 {
		return fmt.Errorf("format/fix case proof binding must occur once")
	}
	return nil
}
