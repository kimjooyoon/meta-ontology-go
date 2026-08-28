package main

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func resolveClaimTuple(sourceFile string, source []byte, file *syntax.File, activityName string) claimResolutionReport {
	report := newClaimResolutionReport(sourceFile, source, activityName)
	var selected *syntax.ActivityDecl
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != activityName {
			continue
		}
		report.Subject.ActivityOccurrences++
		selected = activity
	}
	if report.Subject.ActivityOccurrences != 1 {
		return failClaimResolution(report, "META_BINDING", "RESOLVE_ACTIVITY_CARDINALITY", "CLAIM_ACTIVITY_CARDINALITY_INVALID", "RESTORE_EXACTLY_ONE_CLAIM_ACTIVITY")
	}
	if !selected.ValueProgramPresent {
		return failClaimResolution(report, "CLAIM_RESOLUTION", "OBSERVE_VALUE_PROGRAM", "CLAIM_RESOLUTION_PROGRAM_MISSING", "PROVIDE_CLAIM_RESOLUTION_PROGRAM")
	}
	report.Subject.ValueProgramDigest = claimResolutionDigest([]byte(selected.ValueProgram))
	claim, observed, failed := parseClaimValueProgram(selected.ValueProgram)
	report.Summary.FieldsObserved = observed
	if failed != nil {
		return failClaimResolution(report, "CLAIM_RESOLUTION", "PARSE_CLAIM_RESOLUTION_TUPLE", failed.reason, failed.next)
	}
	report.Decision = claimDecisionObserved
	report.Claim = claim
	report.Summary.ResolutionsObserved = 1
	report.Indicators = buildClaimResolutionIndicators(report, true)
	return report
}

func newClaimResolutionReport(sourceFile string, source []byte, activity string) claimResolutionReport {
	return claimResolutionReport{
		Schema:    claimResolutionSchema,
		Candidate: claimResolutionCandidateID,
		Decision:  claimDecisionFailed,
		Subject: claimResolutionSubject{
			SourceFile: sourceFile, SourceDigest: claimResolutionDigest(source), Activity: activity,
			Binding: "GOOO_ACTIVITY_VALUE_PROGRAM",
		},
		Contract: claimResolutionContract{
			Version: "v1", BaseFields: []string{"state", "stage", "step", "reason", "unknown_class", "next_operation"},
			States: []string{claimStateClosed, claimStateUnknown, claimStateRefuted}, UnknownClassRequired: true,
		},
		Summary:   claimResolutionSummary{FieldsTotal: 6, ResolutionsTotal: 1},
		Authority: claimResolutionAuthority{Source: "GOOO_ACTIVITY_VALUE_PROGRAM"},
	}
}

func failClaimResolution(report claimResolutionReport, stage, step, reason, next string) claimResolutionReport {
	report.Decision = claimDecisionFailed
	report.Claim = claimResolutionClaim{
		State: claimStateRefuted, Stage: claimOptional(stage), Step: claimOptional(step), Reason: reason, NextOperation: next,
	}
	report.Indicators = buildClaimResolutionIndicators(report, false)
	return report
}

func buildClaimResolutionIndicators(report claimResolutionReport, observed bool) []claimResolutionIndicator {
	resolved := 0
	if observed {
		resolved = 1
	}
	return []claimResolutionIndicator{
		{ID: "gooo.metric.claim-resolution.fields.v1", Value: report.Summary.FieldsObserved, Total: 6, Unit: "fields", Class: "DRIVER", Activity: report.Subject.Activity},
		{ID: "gooo.metric.claim-resolution.activity-cardinality.v1", Value: report.Subject.ActivityOccurrences, Total: 1, Unit: "activities", Class: "GUARDRAIL", Activity: report.Subject.Activity},
		{ID: "gooo.metric.claim-resolution.observed.v1", Value: resolved, Total: 1, Unit: "resolutions", Class: "OUTCOME", Activity: report.Subject.Activity},
		{ID: "gooo.metric.claim-resolution.repository-writes.v1", Value: 0, Total: 0, Unit: "writes", Class: "GUARDRAIL", Activity: report.Subject.Activity},
	}
}

func claimResolutionDigest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
