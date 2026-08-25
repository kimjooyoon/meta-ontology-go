package languagedelivery

func toolObligations() []Obligation {
	return bindProfileObligation([]Obligation{
		obligation("TOOL-STRUCTURED-OUTPUT", AudienceToolAuthor, ClassDriver, "consume versioned machine output", rule(SourceUserJourney, EvidenceIndicator, "functional.structured-output", "", 1), "decode-versioned-tool-output", ProofFoundation),
		obligation("TOOL-FORMAT-FIX", AudienceToolAuthor, ClassDriver, "format and plan bounded fixes", rule(SourceConformance, EvidenceSurface, "toolchain-format-fix", "", 1), "consume-format-fix-receipt", ProofCoherence),
		obligation("TOOL-LSP-DIAGNOSTICS", AudienceToolAuthor, ClassDriver, "receive editor diagnostics", rule(SourceLSP, EvidenceLSPCounter, "", "diagnostic_paths", 3), "project-lsp-diagnostics", ProofCoherence),
		obligation("TOOL-LSP-NAVIGATION", AudienceToolAuthor, ClassDriver, "navigate exact editor symbols", rule(SourceLSP, EvidenceLSPCounter, "", "navigation_paths", 3), "project-lsp-navigation", ProofRegression),
		obligation("TOOL-CONFORMANCE-AGGREGATION", AudienceToolAuthor, ClassDriver, "aggregate exact toolchain surfaces", rule(SourceConformance, EvidenceConformance, "", "surfaces_satisfied", 9), "reduce-toolchain-conformance", ProofFoundation),
		obligation("TOOL-CROSS-PLATFORM-RELEASE", AudienceToolAuthor, ClassDriver, "consume replayed release archives", rule(SourceRelease, EvidenceRelease, "", "platform_receipts_and_smokes", 3), "bind-cross-platform-release", ProofRegression),
		obligation("TOOL-GO127-INTEROPERATION", AudienceToolAuthor, ClassDriver, "project Go 1.27 language boundaries", rule(SourceConformance, EvidenceSurface, "language-go-interoperation", "", 1), "consume-go127-interoperation", ProofCoherence),
		obligation("TOOL-PACKAGE-SEMANTIC-MODEL", AudienceToolAuthor, ClassDriver, "inspect package initialization semantics", rule(SourceConformance, EvidenceSurface, "language-package-runtime", "", 1), "consume-package-semantic-model", ProofFoundation),
		obligation("TOOL-DETERMINISTIC-QUERY-MODEL", AudienceToolAuthor, ClassDriver, "query deterministic semantic state", rule(SourceConformance, EvidenceSurface, "language-deterministic-query", "", 1), "consume-deterministic-query-model", ProofCoherence),
		obligation("TOOL-DIAGNOSTIC-PROVENANCE", AudienceToolAuthor, ClassDriver, "trace diagnostics to source and meaning", rule(SourceConformance, EvidenceSurface, "language-diagnostic-provenance", "", 1), "consume-diagnostic-provenance", ProofRegression),
		obligation("TOOL-DEBUGGER", AudienceToolAuthor, ClassDriver, "debug executing Gooo code", missing(), "require-debugger-receipt", ProofCoherence),
		obligation("TOOL-PROFILER", AudienceToolAuthor, ClassDriver, "profile executing Gooo code", missing(), "require-profiler-receipt", ProofRegression),
	})
}
