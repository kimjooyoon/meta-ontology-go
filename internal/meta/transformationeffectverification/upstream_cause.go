package transformationeffectverification

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"

func projectUpstreamCause(target *Report, receipt generation.ReceiptReport) {
	if target == nil || !validUpstreamReceiptReport(receipt) {
		return
	}
	cause := &UpstreamCause{
		ReceiptReportDigest: receipt.ReportDigest,
		Decision:            string(receipt.Decision),
		Reason:              string(receipt.Reason),
		FailureCount:        len(receipt.Failures),
		UnknownCount:        len(receipt.Unknowns),
		Failures:            make([]UpstreamFailure, 0, len(receipt.Failures)),
		Unknowns:            make([]UpstreamUnknown, 0, len(receipt.Unknowns)),
	}
	for _, failure := range receipt.Failures {
		cause.Failures = append(cause.Failures, UpstreamFailure{
			ActionIndicatorID: failure.ActionIndicatorID,
			Decision:          failure.Decision,
			Stage:             failure.Stage,
			Step:              failure.Step,
			Reason:            failure.Reason,
			UnknownClass:      failure.UnknownClass,
			NextOperation:     failure.NextOperation,
			BlockedBy:         cloneStrings(failure.BlockedBy),
		})
	}
	for _, unknown := range receipt.Unknowns {
		cause.Unknowns = append(cause.Unknowns, UpstreamUnknown{
			ActionIndicatorID:   unknown.ActionIndicatorID,
			RequiredIndicatorID: unknown.RequiredIndicatorID,
			Stage:               unknown.Stage,
			Step:                unknown.Step,
			Reason:              string(unknown.Reason),
			UnknownClass:        unknown.UnknownClass,
			NextOperation:       unknown.NextOperation,
			BlockedBy:           cloneStrings(unknown.BlockedBy),
		})
	}
	target.Upstream = cause
}

func validUpstreamReceiptReport(report generation.ReceiptReport) bool {
	if !validUpstreamDigest(report.ReportDigest) ||
		(report.Decision != generation.ReceiptDecisionRefuted && report.Decision != generation.ReceiptDecisionUnknown) ||
		report.Reason == "" || len(report.Failures) == 0 && len(report.Unknowns) == 0 {
		return false
	}
	for _, failure := range report.Failures {
		if failure.ActionIndicatorID == "" || (failure.Decision != "REFUTED" && failure.Decision != "UNKNOWN") ||
			failure.Stage == "" || failure.Step == "" || failure.Reason == "" ||
			failure.NextOperation == "" || failure.BlockedBy == nil ||
			failure.Decision == "UNKNOWN" && failure.UnknownClass == "" {
			return false
		}
	}
	for _, unknown := range report.Unknowns {
		if unknown.ActionIndicatorID == "" || unknown.RequiredIndicatorID == "" ||
			unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" ||
			unknown.UnknownClass == "" || unknown.NextOperation == "" || unknown.BlockedBy == nil {
			return false
		}
	}
	return true
}

func validUpstreamDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !('0' <= character && character <= '9') &&
			!('a' <= character && character <= 'f') {
			return false
		}
	}
	return true
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string{}, values...)
}
