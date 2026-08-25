package languagedelivery

func baseUserObligations() []Obligation {
	return []Obligation{
		obligation("USER-CLI-IDENTITY", AudienceUser, ClassOutcome, "identify the installed Gooo executable", rule(SourceUserJourney, EvidenceJourney, "version-text", "", 1), "observe-installed-language-identity", ProofFoundation),
		obligation("USER-COMMAND-DISCOVERY", AudienceUser, ClassOutcome, "discover the declared command surface", rule(SourceUserJourney, EvidenceIndicator, "functional.declared-commands", "", 1), "project-user-command-surface", ProofFoundation),
		obligation("USER-SYNTAX-CHECK", AudienceUser, ClassOutcome, "check a valid Gooo source", rule(SourceUserJourney, EvidenceJourney, "check-text", "", 1), "execute-user-syntax-check", ProofCoherence),
		obligation("USER-INVALID-DIAGNOSTIC", AudienceUser, ClassOutcome, "reject invalid source with positioned diagnostics", rule(SourceConformance, EvidenceSurface, "language-diagnostic-provenance", "", 1), "project-fail-closed-diagnostic", ProofRegression),
		obligation("USER-SEMANTIC-CHECK", AudienceUser, ClassOutcome, "check semantic meaning without mutation", rule(SourceUserJourney, EvidenceJourney, "semantic-check", "", 1), "execute-user-semantic-check", ProofCoherence),
		obligation("USER-ROUNDTRIP", AudienceUser, ClassOutcome, "roundtrip source with semantic preservation", rule(SourceUserJourney, EvidenceJourney, "roundtrip-json", "", 1), "execute-user-roundtrip", ProofCoherence),
		obligation("USER-RUN-SOURCE", AudienceUser, ClassOutcome, "execute a Gooo source program", rule(SourceExecution, EvidenceExecution, "", "source_executions", 1), "execute-source-activity", ProofFoundation),
		obligation("USER-DETERMINISTIC-RUNTIME-RESULT", AudienceUser, ClassOutcome, "observe a deterministic runtime result", rule(SourceExecution, EvidenceExecution, "", "deterministic_replays", 1), "replay-source-execution-result", ProofCoherence),
		obligation("USER-RUNTIME-FAILURE-DIAGNOSTIC", AudienceUser, ClassOutcome, "observe a fail-closed runtime error", rule(SourceExecution, EvidenceExecution, "", "diagnostic_rejections", 2), "reject-source-runtime-failure", ProofRegression),
		obligation("USER-MULTI-FILE-EXECUTION", AudienceUser, ClassOutcome, "execute a multi-file package", missing(), "require-multi-file-execution", ProofCoherence),
		obligation("USER-EXTERNAL-DEPENDENCY-EXECUTION", AudienceUser, ClassOutcome, "execute a program with an external dependency", missing(), "require-external-dependency-execution", ProofRegression),
		obligation("USER-LANGUAGE-TEST", AudienceUser, ClassOutcome, "run a language-level test", missing(), "require-language-test-receipt", ProofRegression),
	}
}
