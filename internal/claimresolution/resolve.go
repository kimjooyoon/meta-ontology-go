package claimresolution

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Resolve(sourceFile string, source []byte, file *syntax.File, activityName string) Report {
	report := newReport(sourceFile, source, activityName)
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
		return fail(report, "META_BINDING", "RESOLVE_ACTIVITY_CARDINALITY", "CLAIM_ACTIVITY_CARDINALITY_INVALID", "RESTORE_EXACTLY_ONE_CLAIM_ACTIVITY")
	}
	if !selected.ValueProgramPresent {
		return fail(report, "CLAIM_RESOLUTION", "OBSERVE_VALUE_PROGRAM", "CLAIM_RESOLUTION_PROGRAM_MISSING", "PROVIDE_CLAIM_RESOLUTION_PROGRAM")
	}
	report.Subject.ValueProgramDigest = digest([]byte(selected.ValueProgram))
	claim, observed, failed := parseValueProgram(selected.ValueProgram)
	report.Summary.FieldsObserved = observed
	if failed != nil {
		return fail(report, "CLAIM_RESOLUTION", "PARSE_CLAIM_RESOLUTION_TUPLE", failed.reason, failed.next)
	}
	report.Decision = DecisionObserved
	report.Claim = claim
	report.Summary.ResolutionsObserved = 1
	report.Indicators = indicators(report, true)
	return report
}

func newReport(sourceFile string, source []byte, activity string) Report {
	return Report{
		Schema: Schema, Candidate: CandidateID, Decision: DecisionFailed,
		Subject: Subject{SourceFile: sourceFile, SourceDigest: digest(source), Activity: activity, Binding: "GOOO_ACTIVITY_VALUE_PROGRAM"},
		Contract: Contract{
			Version: "v1", BaseFields: []string{"state", "stage", "step", "reason", "unknown_class", "next_operation"},
			States: []string{StateClosed, StateUnknown, StateRefuted}, UnknownClassRequired: true,
		},
		Summary: Summary{FieldsTotal: 6, ResolutionsTotal: 1},
		Authority: Authority{Source: "GOOO_ACTIVITY_VALUE_PROGRAM", CoreMutationAuthorized: false, RepositoryWrites: 0},
	}
}

func fail(report Report, stage, step, reason, next string) Report {
	report.Decision = DecisionFailed
	report.Claim = Claim{State: StateRefuted, Stage: optional(stage), Step: optional(step), Reason: reason, NextOperation: next}
	report.Indicators = indicators(report, false)
	return report
}

func indicators(report Report, observed bool) []Indicator {
	resolved := 0
	if observed {
		resolved = 1
	}
	return []Indicator{
		{ID: "gooo.metric.claim-resolution.fields.v1", Value: report.Summary.FieldsObserved, Total: 6, Unit: "fields", Class: "DRIVER", Activity: report.Subject.Activity},
		{ID: "gooo.metric.claim-resolution.activity-cardinality.v1", Value: report.Subject.ActivityOccurrences, Total: 1, Unit: "activities", Class: "GUARDRAIL", Activity: report.Subject.Activity},
		{ID: "gooo.metric.claim-resolution.observed.v1", Value: resolved, Total: 1, Unit: "resolutions", Class: "OUTCOME", Activity: report.Subject.Activity},
		{ID: "gooo.metric.claim-resolution.repository-writes.v1", Value: 0, Total: 0, Unit: "writes", Class: "GUARDRAIL", Activity: report.Subject.Activity},
	}
}

func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
