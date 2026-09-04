package verify

const compilerSelfImprovement20260903Branch = "agent/compiler-self-improvement-20260903"
const compilerEnvelopeSelfImprovement20260903Branch = "agent/compiler-envelope-self-improvement-20260903"
const compilerEnvelopeSelfImprovementIteration20260904Branch = "agent/compiler-envelope-self-improvement-iteration-20260904"
const compilerEnvelopeSelfImprovementIteration20260904LSPCacheBranch = "agent/compiler-envelope-self-improvement-iteration-20260904-lsp-cache"
const compilerSelfObservationIteration20260904Branch = "agent/compiler-self-observation-iteration-20260904"
const compilerSelfImprovementIteration20260904V5Branch = "agent/compiler-self-improvement-iteration-20260904-v5"
const compilerSelfImprovementIteration20260904V6Branch = "agent/compiler-self-improvement-iteration-20260904-v6"
const compilerSelfImprovementIteration20260904V8Branch = "agent/compiler-self-improvement-iteration-20260904-v8"
const compilerSelfImprovementIteration20260904V9Branch = "agent/compiler-self-improvement-iteration-20260904-v9"
const compilerSelfImprovementIteration20260904V10Branch = "agent/compiler-self-improvement-iteration-20260904-v10"

func init() {
	branchScopeAllowlist[compilerSelfImprovement20260903Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/compiler-self-improvement.yml",
		".github/workflows/transformation-effect.yml",
		"examples/compiler-self-improvement",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/generator/normalize_part01.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v24.json",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/compiler-self-improvement",
	}
	branchScopeAllowlist[compilerEnvelopeSelfImprovement20260903Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/compiler-self-improvement.yml",
		".github/workflows/transformation-effect.yml",
		"examples/compiler-self-improvement",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/generator/types_part01.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v26.json",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/compiler-self-improvement",
	}
	branchScopeAllowlist[compilerEnvelopeSelfImprovementIteration20260904Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/compiler-self-improvement.yml",
		"examples/compiler-self-improvement",
		"internal/generator/normalize_part02.go",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/compiler-self-improvement",
	}
	branchScopeAllowlist[compilerEnvelopeSelfImprovementIteration20260904LSPCacheBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/compiler-self-improvement.yml",
		"examples/compiler-self-improvement",
		"internal/lsp/features_part03.go",
		"internal/lsp/server_part01.go",
		"internal/lsp/server_part01_test.go",
		"internal/lsp/server_part04.go",
		"internal/lsp/state.go",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/compiler-self-improvement",
	}
	branchScopeAllowlist[compilerSelfObservationIteration20260904Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/compiler-self-improvement.yml",
		".github/workflows/self-improvement-observation.yml",
		"cmd/gooo/generate_observation_part04.go",
		"cmd/gooo/generate_pipeline_part04.go",
		"cmd/gooo/main_part01.go",
		"cmd/gooo/observe_part01.go",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/self-improvement-observation",
		"internal/meta/generation/semantic_observation.go",
		"internal/meta/generation/semantic_observation_verify.go",
		"internal/meta/generation/semantic_operation_envelope_generate.go",
		"internal/meta/generation/semantic_operation_envelope_types.go",
		"internal/meta/generation/semantic_operation_envelope_verify.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/compiler-self-improvement",
		"scripts/self-improvement-observation",
	}
	branchScopeAllowlist[compilerSelfImprovementIteration20260904V5Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-adoption.yml",
		"cmd/gooo/main_part01.go",
		"cmd/gooo/self_adoption_adopt_part01.go",
		"cmd/gooo/self_adoption_generation_part01.go",
		"cmd/gooo/self_adoption_proposal_part01.go",
		"examples/self-improvement-observation",
		"internal/meta/generation/semantic_adoption.go",
		"internal/meta/generation/semantic_observation_public.go",
		"internal/meta/generation/semantic_observation_verify.go",
		"internal/meta/generation/semantic_operation_envelope_types.go",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/self-improvement-adoption",
	}
	branchScopeAllowlist[compilerSelfImprovementIteration20260904V6Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-retained-knowledge.yml",
		"cmd/gooo/emit_dispatch.go",
		"cmd/gooo/self_retention_certify_part01.go",
		"cmd/gooo/self_retention_consume_part01.go",
		"cmd/gooo/self_retention_helpers_part01.go",
		"examples/self-improvement-observation",
		"internal/meta/generation/semantic_retention.go",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/self-improvement-retained-knowledge",
	}
	branchScopeAllowlist[compilerSelfImprovementIteration20260904V8Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/generate_pipeline_part03.go",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
	}
	branchScopeAllowlist[compilerSelfImprovementIteration20260904V9Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-retained-knowledge.yml",
		"cmd/gooo/public_generation_part01.go",
		"cmd/gooo/public_generation_part02.go",
		"cmd/gooo/self_retention_certify_part01.go",
		"cmd/gooo/self_retention_consume_part01.go",
		"cmd/gooo/self_retention_helpers_part01.go",
		"examples/self-improvement-observation",
		"internal/meta/generation/semantic_public_generation.go",
		"internal/meta/generation/semantic_retention.go",
		"internal/meta/retentionpolicy",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/self-improvement-public-generation",
		"scripts/self-improvement-retained-knowledge",
		"scripts/self-improvement-retention-policy",
	}
	branchScopeAllowlist[compilerSelfImprovementIteration20260904V10Branch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-public-discovery.yml",
		"cmd/gooo/cli_json_part01.go",
		"cmd/gooo/generate_part01.go",
		"cmd/gooo/generate_pipeline_part03.go",
		"examples/self-improvement-discovery",
		"internal/meta/discoverypolicy",
		"internal/meta/publicdiscovery",
		"internal/verify/scope_compiler_self_improvement_20260903.go",
		"scripts/self-improvement-discovery-policy",
		"scripts/self-improvement-public-discovery",
	}
}
