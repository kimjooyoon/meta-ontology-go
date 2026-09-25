package policycompilation

import (
	"bytes"
	"fmt"
	"go/format"
	"runtime"
)

type GoGuardRendering struct {
	State                string                 `json:"state"`
	Reason               string                 `json:"reason"`
	ActivityID           string                 `json:"activity_id"`
	Toolchain            string                 `json:"toolchain"`
	ExpectedToolchain    string                 `json:"expected_toolchain"`
	InputDigest          string                 `json:"input_digest,omitempty"`
	OutputDigest         string                 `json:"output_digest,omitempty"`
	ReplayDigest         string                 `json:"replay_digest,omitempty"`
	NativeCalls          int                    `json:"native_calls"`
	FixedPoint           bool                   `json:"fixed_point"`
	OutsideEditPreserved bool                   `json:"outside_edit_preserved"`
	Pending              *PolicyRevisionPending `json:"pending,omitempty"`
}

func renderGoGuardPipeline(plan goGuardPipelinePlan, source []byte, raw GoErrorGuardProposal) (GoGuardRendering, string, error) {
	report := GoGuardRendering{
		State: "UNKNOWN", ActivityID: plan.formatActivity,
		Toolchain: runtime.Version(), ExpectedToolchain: plan.toolchain, InputDigest: raw.CandidateDigest,
	}
	if report.Toolchain != report.ExpectedToolchain {
		return declineGoGuardRendering(report, "REFUTED", "GO_RENDER_TOOLCHAIN_PIN_MISMATCH", "", "")
	}
	report.NativeCalls++
	original, err := format.Source(source)
	if err != nil {
		return declineGoGuardRendering(report, "REFUTED", "GO_RENDER_ORIGINAL_SYNTAX_INVALID", "", "")
	}
	if !bytes.Equal(original, source) {
		return declineGoGuardRendering(report, "UNKNOWN", "GO_RENDER_ORIGINAL_NOT_CANONICAL", "UNBOUNDED", "DECLARE_ORIGINAL_CANONICALIZATION")
	}
	report.NativeCalls++
	candidate, err := format.Source([]byte(raw.CandidateSource))
	if err != nil {
		return declineGoGuardRendering(report, "REFUTED", "GO_RENDER_CANDIDATE_SYNTAX_INVALID", "", "")
	}
	report.OutputDigest = DigestBytes(candidate)
	report.OutsideEditPreserved = goGuardCanonicalSpan(source, candidate, raw.EditStart, raw.EditEnd)
	if !report.OutsideEditPreserved {
		return declineGoGuardRendering(report, "REFUTED", "GO_RENDER_OUTSIDE_EDIT_CHANGED", "", "")
	}
	report.NativeCalls++
	replay, err := format.Source(candidate)
	if err != nil || !bytes.Equal(candidate, replay) {
		return declineGoGuardRendering(report, "REFUTED", "GO_RENDER_FIXED_POINT_MISMATCH", "", "")
	}
	report.ReplayDigest, report.FixedPoint = DigestBytes(replay), true
	report.State, report.Reason = "RENDERED", "GOOO_BOUND_CANONICAL_GO_SOURCE"
	return report, string(candidate), nil
}

func goGuardCanonicalSpan(source, candidate []byte, start, end int) bool {
	if start < 0 || end < start || end > len(source) || len(candidate) < start+len(source)-end {
		return false
	}
	return bytes.HasPrefix(candidate, source[:start]) && bytes.HasSuffix(candidate, source[end:])
}

func declineGoGuardRendering(report GoGuardRendering, state, reason, class, next string) (GoGuardRendering, string, error) {
	report.State, report.Reason = state, reason
	if state == "UNKNOWN" {
		pending := revisionPending("CANONICAL_RENDERING", "FORMAT_GENERATED_GUARD", reason, next)
		pending.UnknownClass, pending.BlockedBy = class, []string{}
		report.Pending = &pending
	}
	return report, "", fmt.Errorf("%s: %s", state, reason)
}

func blockedGoGuardRendering(plan goGuardPipelinePlan, raw GoErrorGuardProposal) GoGuardRendering {
	pending := revisionPending("CANONICAL_RENDERING", "FORMAT_GENERATED_GUARD",
		"UPSTREAM_GUARD_NOT_PROPOSED", "RESOLVE_GUARD_CANDIDATE")
	pending.UnknownClass, pending.BlockedBy = "DEPENDENCY_BLOCKED", []string{raw.ActivityID}
	return GoGuardRendering{
		State: "UNKNOWN", Reason: pending.Reason, ActivityID: plan.formatActivity,
		Toolchain: runtime.Version(), ExpectedToolchain: plan.toolchain, Pending: &pending,
	}
}
