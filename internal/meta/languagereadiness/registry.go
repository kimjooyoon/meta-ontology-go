package languagereadiness

const areaLanguage = "LANGUAGE"
const areaToolchain = "TOOLCHAIN"
const areaMeta = "META"
const areaAutonomy = "AUTONOMY"

var obligations = []Obligation{
	{ID: "LANGUAGE-SYNTAX-ROUNDTRIP", Area: areaLanguage, ProofChoice: "FOUNDATION", ConceptID: "language-syntax-roundtrip"},
	{ID: "LANGUAGE-SEMANTIC-MODEL", Area: areaLanguage, ProofChoice: "COHERENCE", ConceptID: "language-semantic-model"},
	{ID: "LANGUAGE-DETERMINISTIC-QUERY", Area: areaLanguage, ProofChoice: "REGRESSION", ConceptID: "language-deterministic-query"},
	{ID: "LANGUAGE-GO-INTEROPERATION", Area: areaLanguage, ProofChoice: "COHERENCE", ConceptID: "language-go-interoperation"},
	{ID: "LANGUAGE-DIAGNOSTIC-PROVENANCE", Area: areaLanguage, ProofChoice: "REGRESSION", ConceptID: "language-diagnostic-provenance"},
	{ID: "LANGUAGE-PACKAGE-RUNTIME", Area: areaLanguage, ProofChoice: "FOUNDATION", ConceptID: "language-package-runtime"},
	{ID: "TOOLCHAIN-CLI", Area: areaToolchain, ProofChoice: "FOUNDATION", ConceptID: "toolchain-cli"},
	{ID: "TOOLCHAIN-FORMAT-FIX", Area: areaToolchain, ProofChoice: "COHERENCE", ConceptID: "toolchain-format-fix"},
	{ID: "TOOLCHAIN-LSP", Area: areaToolchain, ProofChoice: "COHERENCE", ConceptID: "toolchain-lsp"},
	{ID: "TOOLCHAIN-CONFORMANCE", Area: areaToolchain, ProofChoice: "REGRESSION", ConceptID: "toolchain-conformance"},
	{ID: "TOOLCHAIN-CROSS-PLATFORM-RELEASE", Area: areaToolchain, ProofChoice: "REGRESSION", ConceptID: "toolchain-cross-platform-release"},
	{ID: "TOOLCHAIN-EXECUTABLE-USE-CASES", Area: areaToolchain, ProofChoice: "COHERENCE", ConceptID: "toolchain-executable-use-cases"},
	{ID: "META-METRIC-PROGRAM", Area: areaMeta, ProofChoice: "FOUNDATION", ConceptID: "metric-meta-program"},
	{ID: "META-EXECUTABLE-ACTIONABILITY", Area: areaMeta, ProofChoice: "COHERENCE", ConceptID: "executable-actionability"},
	{ID: "META-EFFECT-BOUNDED-OBSERVATION", Area: areaMeta, ProofChoice: "FOUNDATION", ConceptID: "effect-bounded-observation"},
	{ID: "META-MONOTONE-RESOLUTION", Area: areaMeta, ProofChoice: "REGRESSION", ConceptID: "monotone-semantic-resolution"},
	{ID: "META-CAUSAL-FEEDBACK", Area: areaMeta, ProofChoice: "REGRESSION", ConceptID: "causal-feedback-chain"},
	{ID: "META-CONCEPT-GOVERNED-REFACTORING", Area: areaMeta, ProofChoice: "COHERENCE", ConceptID: "concept-governed-refactoring"},
	{ID: "AUTONOMY-CI-SELECTED-REFACTORING", Area: areaAutonomy, ProofChoice: "REGRESSION", ConceptID: "ci-selected-refactoring"},
	{ID: "AUTONOMY-QUANTIFIED-IMPROVEMENT", Area: areaAutonomy, ProofChoice: "COHERENCE", ConceptID: "quantified-improvement"},
	{ID: "AUTONOMY-VERIFIED-TRANSACTION", Area: areaAutonomy, ProofChoice: "REGRESSION", ConceptID: "verified-transformation-transaction"},
	{ID: "AUTONOMY-CHANGE-PROPOSAL", Area: areaAutonomy, ProofChoice: "COHERENCE", ConceptID: "autonomous-change-proposal"},
	{ID: "AUTONOMY-GUARDED-PROMOTION", Area: areaAutonomy, ProofChoice: "FOUNDATION", ConceptID: "guarded-exact-promotion"},
	{ID: "AUTONOMY-ROLLBACK-FIXED-POINT", Area: areaAutonomy, ProofChoice: "REGRESSION", ConceptID: "rollback-fixed-point-recovery"},
}

func Registry() []Obligation {
	return append([]Obligation(nil), obligations...)
}
