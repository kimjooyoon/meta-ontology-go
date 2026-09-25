package rollbackintegrityshadow

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/rollbackfixedpoint"
)

type caseSpec struct {
	name, decision, resolution, mode string
	mutate                           func(*rollbackfixedpoint.Source)
}

func baseSource() rollbackfixedpoint.Source {
	digest := func(value string) string { return "sha256:" + strings.Repeat(value, 64) }
	return rollbackfixedpoint.Source{ExpectedHeadSHA: PredecessorSHA,
		Guard: rollbackfixedpoint.GuardEvidence{FileSHA256: digest("a"),
			ReportDigest: digest("b"), HeadSHA: PredecessorSHA,
			Decision:   guardedpromotion.DecisionFailClosed,
			Reason:     guardedpromotion.ReasonEvidenceUnknown,
			Resolution: guardedpromotion.ResolutionLower,
			Satisfied:  10, Total: 12, Unresolved: 2},
		Transformation: rollbackfixedpoint.TransformationEvidence{
			FileSHA256: digest("c"), LedgerDigest: digest("d"), HeadSHA: PredecessorSHA,
			Decision: "FIXED_POINT", Reason: "EXACT_FIXED_POINT",
			WorkspaceMode: "DISPOSABLE_WORKTREE", WriteBoundary: "SANDBOX_ONLY",
			SourceWorkspaceUnchanged: true}}
}

func authorize(source *rollbackfixedpoint.Source) {
	source.Guard.Decision = guardedpromotion.DecisionAuthorized
	source.Guard.Reason = guardedpromotion.ReasonAuthorized
	source.Guard.Resolution = guardedpromotion.ResolutionExact
	source.Guard.Satisfied, source.Guard.Unresolved = 12, 0
}

func unknownDecision(source *rollbackfixedpoint.Source) { source.Guard.Decision = "FUTURE_DECISION" }
func addEffect(source *rollbackfixedpoint.Source)       { source.Transformation.Effects = 1 }
func mutateSource(source *rollbackfixedpoint.Source) {
	source.Transformation.SourceWorkspaceUnchanged = false
}
func addWrite(source *rollbackfixedpoint.Source) {
	source.RepositoryWrites, source.Guard.RepositoryWrites = 1, 1
}
func authorizeMutation(source *rollbackfixedpoint.Source) {
	source.Guard.RepositoryMutationAuthorized = true
}

func caseSpecs() []caseSpec {
	return []caseSpec{
		{"recover-fail-closed-fixed-point", "PASS", "EXACT", "RECOVERED_FIXED_POINT", nil},
		{"preserve-authorized-terminal", "PASS", "EXACT", "PROMOTION_AUTHORIZED", authorize},
		{"unknown-guard-decision", "FAIL_CLOSED", "LOWER_RESOLUTION", "", unknownDecision},
		{"reject-transformation-effect", "FAIL_CLOSED", "EXACT", "", addEffect},
		{"reject-source-mutation", "FAIL_CLOSED", "EXACT", "", mutateSource},
		{"reject-observer-write", "FAIL_CLOSED", "EXACT", "", addWrite},
		{"reject-mutation-authority", "FAIL_CLOSED", "EXACT", "", authorizeMutation},
	}
}
