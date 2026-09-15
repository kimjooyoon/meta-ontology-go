package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance"
)

func run(cfg config, stdout io.Writer) error {
	if cfg.root == "" || cfg.subjectSHA == "" || cfg.input == "" || cfg.output == "" || cfg.expectCandidate == "" {
		return fmt.Errorf("root, subject-sha, input, output, and expect-candidate are required")
	}
	if err := requireExternalOutput(cfg.root, cfg.output); err != nil {
		return err
	}
	transaction, err := readTransaction(cfg.input)
	if err != nil {
		return err
	}
	report, err := languageassurance.Evaluate(cfg.subjectSHA, transaction)
	if err != nil {
		return err
	}
	if err := languageassurance.ValidateForSubject(report, cfg.subjectSHA); err != nil {
		return err
	}
	if report.CandidateDecision != cfg.expectCandidate {
		return fmt.Errorf("candidate decision = %s, want %s", report.CandidateDecision, cfg.expectCandidate)
	}
	if err := writeReport(cfg.output, report); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout,
		"language-assurance: assurance=%s candidate=%s operating=%d total=%d coverage_bps=%d unresolved=%d digest=%s\n",
		report.AssuranceDecision, report.CandidateDecision, report.Summary.Operating,
		report.Summary.DenominatorTotal, report.Summary.ImplementationCoverageBPS,
		report.Summary.UnresolvedIndicators, report.ReportDigest)
	return err
}
