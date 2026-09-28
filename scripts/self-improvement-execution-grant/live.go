package main

import (
	"fmt"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
)

func runLive(program grant.PolicyProgram, settings options) error {
	input, err := loadInput(settings, program)
	if err != nil {
		return err
	}
	resolution := grant.Evaluate(program, input)
	decision, decisionSource := "", grant.DecisionSourceSystem
	if resolution.SystemEvidence != nil {
		decision = resolution.SystemEvidence.Decision
	}
	report := grant.LiveReport{Schema: grant.Schema, Policy: program.Evidence, Request: input.Request, GrantDecision: decision, DecisionSource: decisionSource, Resolution: resolution, Verification: grant.Verify(program, input, resolution), Metrics: resolution.Metrics}
	if settings.check {
		if err := grant.VerifyGrantResolution(resolution); err != nil || !report.Verification.Verified {
			return fmt.Errorf("live execution grant check failed: resolution=%v verification=%v", err, report.Verification)
		}
		if resolution.Metrics.LiveGrantRequests != 1 || resolution.ExecutionCount != 0 || resolution.ConsumedUses != 0 ||
			resolution.RepositoryWrites != 0 || resolution.LocalTestExecutions != 0 {
			return fmt.Errorf("live system grant crossed execution boundary: %#v", resolution)
		}
	}
	report.Digest = reportDigest(report)
	return writeJSON(settings.outputPath, report)
}

func reportDigest(report grant.LiveReport) string { report.Digest = ""; return digestJSON(report) }
