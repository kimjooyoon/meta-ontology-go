package generation

import (
	"encoding/json"
	"sort"
)

type receiptInput struct {
	PlanDigest string             `json:"plan_digest"`
	Receipts   []OperationReceipt `json:"receipts"`
}

func finishReceiptReport(report ReceiptReport) ReceiptReport {
	if report.Receipts == nil {
		report.Receipts = []OperationReceipt{}
	}
	if report.MissingIndicatorIDs == nil {
		report.MissingIndicatorIDs = []string{}
	}
	if report.UnknownIndicatorIDs == nil {
		report.UnknownIndicatorIDs = []string{}
	}
	if report.RejectedIndicatorIDs == nil {
		report.RejectedIndicatorIDs = []string{}
	}
	sort.Strings(report.MissingIndicatorIDs)
	sort.Strings(report.UnknownIndicatorIDs)
	sort.Strings(report.RejectedIndicatorIDs)
	report.InputDigest = digestJSON(receiptInput{
		PlanDigest: report.PlanDigest,
		Receipts:   report.Receipts,
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
