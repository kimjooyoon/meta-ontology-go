package predecessorresolution

import (
	"fmt"
	"reflect"
)

func Validate(report Report) error {
	if report.Schema != Schema || report.Summary.SearchLimit != SearchLimit ||
		report.Summary.CoordinatesTotal != 10 || report.ReportDigest == "" {
		return fmt.Errorf("resolution contract malformed")
	}
	digest := report.ReportDigest
	report.ReportDigest = ""
	if digestJSON(report) != digest {
		return fmt.Errorf("resolution digest mismatch")
	}
	contiguous := receiptContiguous(report.Attempts)
	expectedIndicators := indicators(report, contiguous)
	if !reflect.DeepEqual(report.Indicators, expectedIndicators) {
		return fmt.Errorf("resolution indicators mismatch")
	}
	if !reflect.DeepEqual(report.Proofs, proofs(report, contiguous)) {
		return fmt.Errorf("resolution proofs mismatch")
	}
	if report.Decision == DecisionResolved {
		if report.Reason != ReasonResolved || report.Selected == nil ||
			report.Summary.CoordinatesCompleted != 10 ||
			report.Summary.BasisPoints != 10000 {
			return fmt.Errorf("resolved coordinates malformed")
		}
	}
	return nil
}

func receiptContiguous(attempts []AttemptReceipt) bool {
	for index := 1; index < len(attempts); index++ {
		if attempts[index-1].ParentSHA != attempts[index].AncestorSHA {
			return false
		}
	}
	return true
}
