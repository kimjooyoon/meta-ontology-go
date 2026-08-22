package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackstate"
)

func digestBytes(values ...[]byte) string {
	hash := sha256.New()
	for _, value := range values {
		_, _ = fmt.Fprintf(hash, "%d:", len(value))
		_, _ = hash.Write(value)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

func rawDigest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func newSemanticReceipt(report, replay feedbackstate.Report, predecessorDigest, inputDigest string) semanticReceipt {
	receipt := semanticReceipt{
		Schema: semanticReceiptSchema, Report: report,
		PredecessorSelectionReceiptDigest: predecessorDigest,
		InputDigest: inputDigest, ReplayReportDigest: replay.ReportDigest,
		ReplayVerified: report.ReportDigest == replay.ReportDigest,
	}
	encoded, _ := json.Marshal(receipt)
	receipt.ReceiptDigest = rawDigest(encoded)
	return receipt
}

func marshalReceipt(receipt semanticReceipt) ([]byte, error) {
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
