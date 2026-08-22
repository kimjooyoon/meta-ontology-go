package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func run(cfg config) (bool, error) {
	if cfg.root == "" || cfg.input == "" || cfg.report == "" {
		return false, fmt.Errorf("root, input, and report are required")
	}
	outside, err := outsideRoot(cfg.root, cfg.report)
	if err != nil {
		return false, err
	}
	if !outside {
		return false, fmt.Errorf("predecessor receipt must be outside the repository root")
	}
	input, err := readInput(cfg.input)
	if err != nil {
		return false, err
	}
	report, err := feedbackpredecessor.Select(input)
	if err != nil {
		return false, err
	}
	replay, err := feedbackpredecessor.Select(input)
	if err != nil {
		return false, err
	}
	if report.ReportDigest != replay.ReportDigest {
		return false, fmt.Errorf("predecessor replay digest mismatch")
	}
	if cfg.expectedDigest != "" && report.ReportDigest != cfg.expectedDigest {
		return false, fmt.Errorf("predecessor report digest does not match expected digest")
	}
	output, err := newReceipt(report, replay.ReportDigest, cfg.expectedDigest)
	if err != nil {
		return false, err
	}
	data, err := marshalReceipt(output)
	if err != nil {
		return false, err
	}
	if err := writeExclusive(cfg.report, data); err != nil {
		return false, err
	}
	fmt.Printf("feedback-predecessor: decision=%s reason=%s digest=%s\n",
		report.Decision, report.Reason, report.ReportDigest)
	return cfg.check && report.Decision == feedbackpredecessor.DecisionFailClosed, nil
}
