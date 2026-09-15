package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackstate"
)

const (
	predecessorReceiptSchema = "gooo/meta-feedback-predecessor-receipt/v1"
	semanticReceiptSchema    = "gooo/meta-feedback-semantic-snapshot-receipt/v1"
)

type config struct {
	root               string
	input              string
	predecessorReceipt string
	report             string
	expectedDigest     string
	check              bool
}

type predecessorReceipt struct {
	Schema             string                     `json:"schema"`
	Report             feedbackpredecessor.Report `json:"report"`
	ReplayReportDigest string                     `json:"replay_report_digest"`
	ExpectedDigest     string                     `json:"expected_digest,omitempty"`
	ReplayVerified     bool                       `json:"replay_verified"`
	RepositoryWrites   int                        `json:"repository_writes"`
	ReceiptDigest      string                     `json:"receipt_digest"`
}

type semanticReceipt struct {
	Schema                            string               `json:"schema"`
	Report                            feedbackstate.Report `json:"report"`
	PredecessorSelectionReceiptDigest string               `json:"predecessor_selection_receipt_digest"`
	InputDigest                       string               `json:"input_digest"`
	ReplayReportDigest                string               `json:"replay_report_digest"`
	ReplayVerified                    bool                 `json:"replay_verified"`
	RepositoryWrites                  int                  `json:"repository_writes"`
	ReceiptDigest                     string               `json:"receipt_digest"`
}
