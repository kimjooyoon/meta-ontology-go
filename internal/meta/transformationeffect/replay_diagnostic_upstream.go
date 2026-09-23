package transformationeffect

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const replayDiagnosticUpstreamSchema = "gooo/transformation-effect-upstream-receipt/v1"

type replayDiagnosticContext struct {
	Cause  error
	Report generation.ReceiptReport
}

func (err *replayDiagnosticContext) Error() string {
	return err.Cause.Error()
}

func (err *replayDiagnosticContext) Unwrap() error {
	return err.Cause
}

type ReplayDiagnosticFailure struct {
	ActionIndicatorID string   `json:"action_indicator_id"`
	Decision          string   `json:"decision"`
	Stage             string   `json:"stage"`
	Step              string   `json:"step"`
	Reason            string   `json:"reason"`
	NextOperation     string   `json:"next_operation"`
	BlockedBy         []string `json:"blocked_by"`
}

type ReplayDiagnosticUpstream struct {
	Schema              string                     `json:"schema"`
	BaseSHA             string                     `json:"base_sha"`
	HeadSHA             string                     `json:"head_sha"`
	PlanDigest          string                     `json:"plan_digest"`
	InputDigest         string                     `json:"input_digest"`
	ReportDigest        string                     `json:"report_digest"`
	Decision            string                     `json:"decision"`
	Reason              string                     `json:"reason"`
	PromotionAuthorized bool                       `json:"promotion_authorized"`
	Failures            []ReplayDiagnosticFailure  `json:"failures"`
	Unknowns            []CausalUnknownRecord      `json:"unknowns"`
	CausalDigest        string                     `json:"causal_digest"`
}

func newReplayDiagnosticUpstream(report generation.ReceiptReport) *ReplayDiagnosticUpstream {
	if report.PromotionAuthorized {
		return nil
	}
	projection, err := BuildCausalUnknownProjection(report)
	if err != nil {
		return nil
	}
	failures := make([]ReplayDiagnosticFailure, 0, len(report.Failures))
	for _, failure := range report.Failures {
		failures = append(failures, ReplayDiagnosticFailure{
			ActionIndicatorID: failure.ActionIndicatorID,
			Decision:          failure.Decision,
			Stage:             failure.Stage,
			Step:              failure.Step,
			Reason:            failure.Reason,
			NextOperation:     failure.NextOperation,
			BlockedBy:         append([]string{}, failure.BlockedBy...),
		})
	}
	unknowns := append([]CausalUnknownRecord{}, projection.Records...)
	return &ReplayDiagnosticUpstream{
		Schema:              replayDiagnosticUpstreamSchema,
		BaseSHA:             report.BaseSHA,
		HeadSHA:             report.HeadSHA,
		PlanDigest:          report.PlanDigest,
		InputDigest:         report.InputDigest,
		ReportDigest:        report.ReportDigest,
		Decision:            string(report.Decision),
		Reason:              string(report.Reason),
		PromotionAuthorized: report.PromotionAuthorized,
		Failures:            failures,
		Unknowns:            unknowns,
		CausalDigest:        projection.Digest,
	}
}

func applyReplayDiagnosticUpstream(diagnostic *ReplayDiagnostic, upstream *ReplayDiagnosticUpstream) {
	if upstream == nil {
		return
	}
	blockedBy := make([]string, 0)
	nextOperations := make(map[string]bool)
	for _, unknown := range upstream.Unknowns {
		if unknown.UnknownClass != generation.ReceiptUnknownClassDependencyBlocked {
			continue
		}
		blockedBy = append(blockedBy, unknown.BlockedBy...)
		nextOperations[unknown.NextOperation] = true
	}
	if len(blockedBy) == 0 {
		return
	}
	sort.Strings(blockedBy)
	unique := blockedBy[:0]
	for _, item := range blockedBy {
		if len(unique) == 0 || unique[len(unique)-1] != item {
			unique = append(unique, item)
		}
	}
	diagnostic.UnknownClass = generation.ReceiptUnknownClassDependencyBlocked
	diagnostic.Reason = "UPSTREAM_RECEIPT_DEPENDENCY_BLOCKED"
	diagnostic.BlockedBy = unique
	if len(nextOperations) == 1 {
		for operation := range nextOperations {
			diagnostic.NextOperation = operation
		}
		return
	}
	diagnostic.NextOperation = "restore-upstream-evidence"
}
