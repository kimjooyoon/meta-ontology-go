package generation

import (
	"encoding/json"
	"sort"
)

type receiptInput struct {
	PlanDigest                    string               `json:"plan_digest"`
	IndicatorDecisionLedgerDigest string               `json:"indicator_decision_ledger_digest,omitempty"`
	IndicatorDecisionLedgerCount  int                  `json:"indicator_decision_ledger_count"`
	Receipts                      []OperationReceipt   `json:"receipts"`
	Failures                      []ObservationFailure `json:"failures"`
	Unknowns                      []ReceiptUnknown     `json:"unknowns"`
}

func finishReceiptReport(report ReceiptReport) ReceiptReport {
	if report.Receipts == nil {
		report.Receipts = []OperationReceipt{}
	}
	report.Failures = normalizeObservationFailures(report.Failures)
	if report.MissingIndicatorIDs == nil {
		report.MissingIndicatorIDs = []string{}
	}
	if report.UnknownIndicatorIDs == nil {
		report.UnknownIndicatorIDs = []string{}
	}
	if report.RejectedIndicatorIDs == nil {
		report.RejectedIndicatorIDs = []string{}
	}
	report.Unknowns = normalizeReceiptUnknowns(report.Unknowns)
	sort.Strings(report.MissingIndicatorIDs)
	sort.Strings(report.UnknownIndicatorIDs)
	sort.Strings(report.RejectedIndicatorIDs)
	report.InputDigest = digestJSON(receiptInput{
		PlanDigest:                    report.PlanDigest,
		IndicatorDecisionLedgerDigest: report.IndicatorDecisionLedgerDigest,
		IndicatorDecisionLedgerCount:  report.IndicatorDecisionLedgerCount,
		Receipts:                      report.Receipts, Failures: report.Failures, Unknowns: report.Unknowns,
	})
	report.ReportDigest, report.ReplayDigest = "", ""
	report.ReportDigest = digestJSON(report)
	report.ReplayDigest = digestPair(report.InputDigest, report.ReportDigest)
	return report
}

func EncodeReceiptReport(report ReceiptReport) ([]byte, error) {
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}
