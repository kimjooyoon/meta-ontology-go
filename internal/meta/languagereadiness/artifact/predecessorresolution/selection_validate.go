package predecessorresolution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func validateSelectionReport(report predecessorselection.Report) error {
	if report.Schema != predecessorselection.Schema || report.ReportDigest == "" {
		return fmt.Errorf("selection contract malformed")
	}
	digest := report.ReportDigest
	report.ReportDigest = ""
	if digestJSON(report) != digest {
		return fmt.Errorf("selection digest mismatch")
	}
	if report.Decision == predecessorselection.DecisionSelected {
		if report.Reason != predecessorselection.ReasonSelected || report.Selected == nil {
			return fmt.Errorf("selected decision malformed")
		}
		return nil
	}
	if report.Decision != predecessorselection.DecisionFailClosed ||
		report.Selected != nil || !knownFailureReason(report.Reason) {
		return fmt.Errorf("fail-closed decision malformed")
	}
	return nil
}

func knownFailureReason(reason string) bool {
	switch reason {
	case predecessorselection.ReasonNotFound,
		predecessorselection.ReasonUnbound,
		predecessorselection.ReasonFailed,
		predecessorselection.ReasonExpired,
		predecessorselection.ReasonInvalid,
		predecessorselection.ReasonAmbiguous,
		predecessorselection.ReasonWriteEffect:
		return true
	default:
		return false
	}
}
