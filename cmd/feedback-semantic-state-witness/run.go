package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackstate"
)

func run(cfg config) (bool, error) {
	if cfg.root == "" || cfg.input == "" || cfg.predecessorReceipt == "" || cfg.report == "" {
		return false, fmt.Errorf("root, input, predecessor-receipt, and report are required")
	}
	outside, err := outsideRoot(cfg.root, cfg.report)
	if err != nil {
		return false, err
	}
	if !outside {
		return false, fmt.Errorf("semantic receipt must be outside the repository root")
	}
	input, inputBytes, err := readJSON[feedbackpredecessor.Input](cfg.input)
	if err != nil {
		return false, err
	}
	predecessor, predecessorBytes, err := readJSON[predecessorReceipt](cfg.predecessorReceipt)
	if err != nil {
		return false, err
	}
	bound, err := bindSemanticInput(input, predecessor)
	if err != nil {
		return false, err
	}
	report, replay := feedbackstate.Evaluate(bound), feedbackstate.Evaluate(bound)
	if report.ReportDigest != replay.ReportDigest {
		return false, fmt.Errorf("semantic snapshot replay digest mismatch")
	}
	if cfg.expectedDigest != "" && report.ReportDigest != cfg.expectedDigest {
		return false, fmt.Errorf("semantic report digest does not match expected digest")
	}
	receipt := newSemanticReceipt(report, replay, predecessor.ReceiptDigest,
		digestBytes(inputBytes, predecessorBytes))
	data, err := marshalReceipt(receipt)
	if err != nil {
		return false, err
	}
	if err := writeExclusive(cfg.report, data); err != nil {
		return false, err
	}
	fmt.Printf("feedback-semantic-state: decision=%s reason=%s digest=%s\n",
		report.Decision, report.Reason, report.ReportDigest)
	return cfg.check && report.Decision != "READY", nil
}
