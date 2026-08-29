package verify

const baseRepairRepositoryProjectionBranch = "agent/base-repair-repository-projection"

func init() {
	branchScopeAllowlist[baseRepairRepositoryProjectionBranch] = []string{
		".github/workflows/ci.yml",
		".github/workflows/causal-ci-selection.yml",
		"bootstrap/function-extractor",
		"bootstrap/line-density-rewriter",
		"bootstrap/logical-split-planner",
		"examples/language-semantic-model/corpus.json",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/repository-projection-repair/main.gooo",
		"examples/self-improvement/main.gooo",
		"examples/source-splitter-conformance/contract.json",
		"internal/meta/actionability/registry.go",
		"internal/meta/artifactcoverage",
		"internal/meta/functionextractorrecipe",
		"internal/meta/generation",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/metabinding",
		"internal/meta/metricstrategy/proposal/conformance.go",
		"internal/meta/operationconformance",
		"internal/meta/repositoryprojection/extractor",
		"internal/meta/sourcepolicy",
		"scripts/authorized-write-set/input_model.go",
		"scripts/meta-execution",
		"scripts/meta-receipts",
		"scripts/meta-summary",
		"scripts/operation-artifact-ci/build-observations.sh",
		"scripts/self-improvement-contract/contract_test.go",
		"scripts/semantic-conformance.sh",
		"scripts/source-splitter",
		"scripts/verify",
		"internal/verify/scope_base_repair_repository_projection.go",
		"internal/verify/scope_base_repair_repository_projection_test.go",
		".github/agent-scope-table.md",
		".github/ci-governance.json",
	}
}
