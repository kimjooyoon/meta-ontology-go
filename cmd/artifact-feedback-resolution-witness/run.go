package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactfeedback"
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
		return false, fmt.Errorf("resolution receipt must be outside the repository root")
	}
	input, err := readInput(cfg.input)
	if err != nil {
		return false, err
	}
	report, err := artifactfeedback.EvaluateWithResolution(input)
	if err != nil {
		return false, err
	}
	replay, err := artifactfeedback.EvaluateWithResolution(input)
	if err != nil {
		return false, err
	}
	if report.ReportDigest != replay.ReportDigest {
		return false, fmt.Errorf("resolution replay digest mismatch")
	}
	if cfg.expectedDigest != "" && report.ReportDigest != cfg.expectedDigest {
		return false, fmt.Errorf("resolution report digest does not match expected digest")
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
	fmt.Printf("artifact-feedback-resolution: decision=%s from=%s to=%s descents=%d digest=%s\n",
		report.Decision, report.FromResolution, report.ToResolution,
		report.Descents, report.ReportDigest)
	return cfg.check && report.Decision == "FAIL_CLOSED", nil
}
