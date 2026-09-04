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
	report := grant.LiveReport{Schema: grant.Schema, Policy: program.Evidence, Request: input.Request, GrantDecision: settings.decision, DecisionSource: settings.decisionSource, Resolution: resolution, Verification: grant.Verify(program, input, resolution), Metrics: resolution.Metrics}
	if settings.check {
		if err := grant.VerifyGrantResolution(resolution); err != nil || !report.Verification.Verified {
			return fmt.Errorf("live execution grant check failed: resolution=%v verification=%v", err, report.Verification)
		}
		if settings.decision == "" && (resolution.Decision != grant.DecisionUnknown || resolution.Reason != grant.ReasonMissingDecision || resolution.Unknown == nil || resolution.Unknown.BlockedBy != "explicit_execution_grant_decision" || resolution.Metrics.LiveGrantRequests != 1 || resolution.Metrics.LiveGrants != 0 || resolution.ExecutionCount != 0 || resolution.ConsumedUses != 0) {
			return fmt.Errorf("live no-decision grant request was not UNKNOWN: %#v", resolution)
		}
	}
	report.Digest = reportDigest(report)
	return writeJSON(settings.outputPath, report)
}

func reportDigest(report grant.LiveReport) string { report.Digest = ""; return digestJSON(report) }
