package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"

const receiptSchema = "gooo/meta-feedback-predecessor-receipt/v1"

type receipt struct {
	Schema             string                     `json:"schema"`
	Report             feedbackpredecessor.Report `json:"report"`
	ReplayReportDigest string                     `json:"replay_report_digest"`
	ExpectedDigest     string                     `json:"expected_digest,omitempty"`
	ReplayVerified     bool                       `json:"replay_verified"`
	RepositoryWrites   int                        `json:"repository_writes"`
	ReceiptDigest      string                     `json:"receipt_digest"`
}
