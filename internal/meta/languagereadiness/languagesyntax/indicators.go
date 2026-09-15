package languagesyntax

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"

func indicators(report Report, drift int) []Indicator {
	s := report.Summary
	return []Indicator{
		metric("readiness-bps", "OUTCOME", "COHERENCE", report.Resolution, s.ReadinessBPS, 10000),
		metric("executed-cases", "DRIVER", "FOUNDATION", report.Resolution, s.Executed, totalCases),
		metric("valid-corpus-files", "DRIVER", "FOUNDATION", report.Resolution, s.ValidCases, validCases),
		metric("invalid-fixtures", "DRIVER", "REGRESSION", report.Resolution, s.InvalidCases, invalidCases),
		metric("ast-shape-replays", "DRIVER", "COHERENCE", report.Resolution, s.ASTReplays, validCases),
		metric("canonical-byte-replays", "DRIVER", "COHERENCE", report.Resolution, s.ByteReplays, validCases),
		metric("semantic-hash-replays", "DRIVER", "COHERENCE", report.Resolution, s.SemanticReplays, validCases),
		metric("get-put-laws", "DRIVER", "COHERENCE", report.Resolution, s.GetPutLaws, validCases),
		metric("put-get-laws", "DRIVER", "COHERENCE", report.Resolution, s.PutGetLaws, validCases),
		metric("diagnostic-rejections", "DRIVER", "REGRESSION", report.Resolution, s.DiagnosticRejections, invalidCases),
		metric("unregistered-gooo.guardrail", "GUARDRAIL", "FOUNDATION", report.Resolution, s.UnregisteredGooo, 0),
		metric("missing-registered.guardrail", "GUARDRAIL", "FOUNDATION", report.Resolution, s.MissingRegistered, 0),
		metric("unresolved.guardrail", "GUARDRAIL", "FOUNDATION", report.Resolution, s.Unresolved, 0),
		metric("repository-writes.guardrail", "GUARDRAIL", "REGRESSION", report.Resolution, report.RepositoryWrites, 0),
		metric("mutation-authority.guardrail", "GUARDRAIL", "REGRESSION", report.Resolution, boolInt(report.MutationAuthorized), 0),
		metric("registry-drift.guardrail", "GUARDRAIL", "FOUNDATION", report.Resolution, drift, 0),
	}
}

func proofs(report Report, drift int) []Proof {
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-versioned-complete-gooo-corpus", EvidenceDigest: digestJSON(report.Source),
			Passed: report.Source.ConceptBound && report.Source.ObservationKnown && drift == 0 && report.Summary.UnregisteredGooo == 0 && report.Summary.MissingRegistered == 0},
		{Choice: "COHERENCE", MetaOperation: "replay-ast-bytes-semantics-and-lens-laws", EvidenceDigest: digestJSON(report.Cases),
			Passed: report.Summary.ValidCases == validCases},
		{Choice: "REGRESSION", MetaOperation: "reject-invalid-syntax-with-zero-effects", EvidenceDigest: digestJSON(report.Summary),
			Passed: report.Summary.InvalidCases == invalidCases && report.RepositoryWrites == 0 && !report.MutationAuthorized},
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func caseStatus(item CaseResult) string {
	if item.Evidence.ObservedDecision == replay.DecisionUnknown {
		return "UNRESOLVED"
	}
	if item.Evidence.ObservedDecision == item.Definition.ExpectedDecision {
		return "SATISFIED"
	}
	return "NOT_SATISFIED"
}
