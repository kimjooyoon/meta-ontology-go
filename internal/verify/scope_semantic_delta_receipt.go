package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-20-semantic-delta-receipt"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/semantic-delta-receipt.yml",
		"cmd/semantic-delta-receipt-witness",
		"docs/language/semantic-delta-receipt.md",
		"examples/semantic-delta-receipt",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/language-semantic-model/corpus.json",
		"examples/toolchain-conformance/corpus.json",
		"internal/meta/languageassurance/semanticdeltareceipt",
		"internal/meta/languageassurance/semanticdeltareceiptconsumer",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/languagereadiness/languagesemantic",
		"internal/meta/languagereadiness/languagesemanticbinding",
		"internal/meta/languagereadiness/toolchainconformance",
		"internal/verify/scope_semantic_delta_receipt.go",
	}
}
