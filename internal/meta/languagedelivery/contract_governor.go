package languagedelivery

func governorObligations() []Obligation {
	return []Obligation{
		governor("GOVERNOR-METRIC-PROGRAM", "META-METRIC-PROGRAM", "bind every indicator to a meta program", "bind-indicator-meta-program", ProofFoundation),
		governor("GOVERNOR-EXECUTABLE-ACTIONABILITY", "META-EXECUTABLE-ACTIONABILITY", "resolve blocking indicators to executors", "resolve-indicator-executor", ProofCoherence),
		governor("GOVERNOR-EFFECT-BOUNDED-OBSERVATION", "META-EFFECT-BOUNDED-OBSERVATION", "bound observer effects", "preserve-read-only-semantic-state", ProofFoundation),
		governor("GOVERNOR-MONOTONE-RESOLUTION", "META-MONOTONE-RESOLUTION", "lower unknown semantic resolution", "lower-semantic-resolution", ProofRegression),
		governor("GOVERNOR-CAUSAL-FEEDBACK", "META-CAUSAL-FEEDBACK", "select causal feedback evidence", "select-predecessor-semantic-state", ProofRegression),
		governor("GOVERNOR-CONCEPT-REFACTORING", "META-CONCEPT-GOVERNED-REFACTORING", "authorize refactoring through concepts", "bind-refactoring-concept", ProofCoherence),
		governor("GOVERNOR-CI-SELECTED-REFACTORING", "AUTONOMY-CI-SELECTED-REFACTORING", "select bounded refactoring in CI", "compact-obvious-lines", ProofRegression),
		governor("GOVERNOR-QUANTIFIED-IMPROVEMENT", "AUTONOMY-QUANTIFIED-IMPROVEMENT", "compare fixed readiness coordinates", "compare-readiness-receipts", ProofCoherence),
		governor("GOVERNOR-VERIFIED-TRANSACTION", "AUTONOMY-VERIFIED-TRANSACTION", "verify transformation transactions", "verify-transformation-transaction", ProofRegression),
		governor("GOVERNOR-CHANGE-PROPOSAL", "AUTONOMY-CHANGE-PROPOSAL", "promote verified change proposals", "promote-verified-change-proposal", ProofCoherence),
		governor("GOVERNOR-GUARDED-PROMOTION", "AUTONOMY-GUARDED-PROMOTION", "guard exact promotion capability", "bind-guarded-capability-foundation", ProofFoundation),
		governor("GOVERNOR-ROLLBACK-FIXED-POINT", "AUTONOMY-ROLLBACK-FIXED-POINT", "recover a non-authorizing fixed point", "recover-guarded-fixed-point", ProofRegression),
	}
}

func governor(id, readinessID, outcome, operation string, proof ProofChoice) Obligation {
	return obligation(id, AudienceGovernor, ClassGuardrail, outcome,
		rule(SourceReadiness, EvidenceReadiness, readinessID, "", 1), operation, proof)
}
