package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactcoverage"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactfeedback"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/semanticresolution"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.coverage == "" || cfg.provenance == "" || cfg.output == "" {
		return fmt.Errorf("root, coverage, provenance, and output are required")
	}
	outside, err := outsideRoot(cfg.root, cfg.output)
	if err != nil {
		return err
	}
	if !outside {
		return fmt.Errorf("resolution input must be outside the repository root")
	}
	var coverage artifactcoverage.Report
	if err := readJSON(cfg.coverage, &coverage); err != nil {
		return err
	}
	var provenance provenanceEnvelope
	if err := readJSON(cfg.provenance, &provenance); err != nil {
		return err
	}
	if err := validateSources(coverage, provenance, cfg.ciConclusion); err != nil {
		return err
	}
	input := artifactfeedback.ResolutionInput{
		Feedback: artifactfeedback.Input{
			Coverage: coverage, CoverageReplayDigest: coverage.ReportDigest,
			Cycle: artifactfeedback.CycleObservation{
				Schema: artifactfeedback.CycleSchema, HeadSHA: provenance.HeadSHA,
				Status: provenance.Decision, CIConclusion: cfg.ciConclusion,
				EnvelopeDigest: provenance.EnvelopeDigest,
				ReplayDigest:   provenance.ReplayDigest,
			},
			RepositoryWrites: coverage.Summary.RepositoryWrites,
		},
		CurrentResolution: semanticresolution.ResolutionExactOperation,
	}
	report, err := artifactfeedback.EvaluateWithResolution(input)
	if err != nil {
		return err
	}
	if report.Decision == "FAIL_CLOSED" {
		return fmt.Errorf("generated resolution input fails closed: %s", report.Reason)
	}
	return writeJSONExclusive(cfg.output, input)
}

func validateSources(coverage artifactcoverage.Report, provenance provenanceEnvelope, conclusion string) error {
	if coverage.CommitSHA == "" || coverage.CommitSHA != provenance.HeadSHA {
		return fmt.Errorf("coverage and provenance heads do not match")
	}
	if provenance.SchemaVersion == "" || provenance.Decision != "BOUND" ||
		provenance.Reason == "" || conclusion != "success" {
		return fmt.Errorf("self-improvement provenance is not bound")
	}
	if !validBareDigest(provenance.EnvelopeDigest) ||
		!validBareDigest(provenance.ReplayDigest) {
		return fmt.Errorf("self-improvement provenance digest is malformed")
	}
	return nil
}
