package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactfeedback"

const receiptSchema = "gooo/meta-artifact-feedback-resolution-receipt/v1"

type receipt struct {
	Schema             string                            `json:"schema"`
	Report             artifactfeedback.ResolutionReport `json:"report"`
	ReplayReportDigest string                            `json:"replay_report_digest"`
	ExpectedDigest     string                            `json:"expected_digest,omitempty"`
	ReplayVerified     bool                              `json:"replay_verified"`
	RepositoryWrites   int                               `json:"repository_writes"`
	ReceiptDigest      string                            `json:"receipt_digest"`
}
