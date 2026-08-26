package rollbackfixedpoint

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"

func Coordinates(source Source) []Coordinate {
	guard := source.Guard
	transformation := source.Transformation
	authorized := authorized(guard)
	recoverable := recoverable(guard)
	terminalKnown := guard.Decision == guardedpromotion.DecisionAuthorized ||
		guard.Decision == guardedpromotion.DecisionFailClosed ||
		guard.Decision == guardedpromotion.DecisionDenied
	fixedPoint := transformation.Decision == "FIXED_POINT" &&
		transformation.Reason == "EXACT_FIXED_POINT"
	return []Coordinate{
		coordinate("exact-expected-subject", "FOUNDATION", validSHA(source.ExpectedHeadSHA),
			validSHA(source.ExpectedHeadSHA)),
		coordinate("canonical-guard-receipt", "FOUNDATION", guard.HeadSHA != "",
			source.CollectionError == "" && validDigest(guard.FileSHA256) &&
				validDigest(guard.ReportDigest)),
		coordinate("canonical-transformation-ledger", "FOUNDATION", transformation.HeadSHA != "",
			source.CollectionError == "" && validDigest(transformation.FileSHA256) &&
				validDigest(transformation.LedgerDigest)),
		coordinate("exact-subject-link", "COHERENCE", guard.HeadSHA != "" &&
			transformation.HeadSHA != "", guard.HeadSHA == source.ExpectedHeadSHA &&
			transformation.HeadSHA == source.ExpectedHeadSHA),
		coordinate("accepted-terminal-path", "COHERENCE", terminalKnown, authorized || recoverable),
		coordinate("recovery-fixed-point", "COHERENCE", terminalKnown,
			authorized || recoverable && fixedPoint),
		coordinate("recovery-effect-boundary", "REGRESSION", terminalKnown,
			authorized || recoverable && transformation.Effects == 0),
		coordinate("source-workspace-boundary", "REGRESSION", transformation.HeadSHA != "",
			transformation.SourceWorkspaceUnchanged && transformation.WriteBoundary == "SANDBOX_ONLY"),
		coordinate("observer-write-boundary", "REGRESSION", guard.HeadSHA != "",
			source.RepositoryWrites == 0 && guard.RepositoryWrites == 0),
		coordinate("mutation-authority-boundary", "REGRESSION", guard.HeadSHA != "",
			!guard.RepositoryMutationAuthorized && !transformation.PromotionAuthorized),
	}
}

func authorized(value GuardEvidence) bool {
	return value.Decision == guardedpromotion.DecisionAuthorized &&
		value.Resolution == guardedpromotion.ResolutionExact && value.Satisfied == 12 &&
		value.Total == 12 && value.Unresolved == 0
}

func recoverable(value GuardEvidence) bool {
	return value.Decision == guardedpromotion.DecisionFailClosed &&
		value.Reason == guardedpromotion.ReasonEvidenceUnknown &&
		value.Resolution == guardedpromotion.ResolutionLower && value.Unresolved > 0
}

func coordinate(id, choice string, observed, passed bool) Coordinate {
	if !observed {
		return Coordinate{id, choice, "UNRESOLVED", "COORDINATE_EVIDENCE_UNKNOWN"}
	}
	if passed {
		return Coordinate{id, choice, "SATISFIED", "COORDINATE_EXACTLY_PROVEN"}
	}
	return Coordinate{id, choice, "NOT_SATISFIED", "COORDINATE_REJECTED"}
}
