# Explicit agent scope registry

This is the registry for the exact branch keys in
`internal/verify/scope.go`. A branch name is never interpreted as a wildcard;
unknown `agent/*` names fail closed. Directory suffixes below describe the
allowed repository prefix and do not create additional branch aliases.

| Branch | Allowed path prefix |
| --- | --- |
| `agent/analyzer` | `internal/analyzer/**` |
| `agent/analyzer-contract` | `internal/analyzer/**` |
| `agent/bidir` | `internal/bidir/**` |
| `agent/bidir-followup` | `internal/bidir/**` |
| `agent/bidir-research` | `docs/research/bidirectional.md` |
| `agent/bidirectional-experiment-contract` | `docs/research/bidirectional.md` |
| `agent/bidirectional-property-matrix` | `docs/research/bidirectional.md` |
| `agent/bidirectional-research` | `docs/research/bidirectional.md` |
| `agent/cache` | `internal/cache/**` |
| `agent/cache-experiment-followup` | `docs/research/cache.md` |
| `agent/cache-research` | `docs/research/cache.md` |
| `agent/ci-evidence-contract` | `.github/**`, `scripts/**`, `internal/verify/**` |
| `agent/ci-ownership-audit` | `.github/**`, `scripts/**`, `internal/verify/**` |
| `agent/ci-ownership-audit-current` | `.github/**`, `scripts/**`, `internal/verify/**` |
| `agent/ci-ownership-audit-current2` | `.github/**`, `scripts/**`, `internal/verify/**` |
| `agent/ci-alias-refresh` | `.github/**`, `scripts/**`, `internal/verify/**` |
| `agent/ci-generator-current7` | `.github/**`, `scripts/**`, `internal/verify/**` |
| `agent/ci-scope-triage` | `.github/**`, `scripts/**`, `internal/verify/**` |
| `agent/ci-workflow` | `.github/**`, `scripts/**`, `internal/verify/**` |
| `agent/ci-workflow-stage` | `.github/**`, `scripts/**`, `internal/verify/**` |
| `agent/cli` | `cmd/gooo/**` |
| `agent/cli-bootstrap-contract` | `cmd/gooo/**` |
| `agent/cli-check` | `cmd/gooo/**` |
| `agent/cli-check-current` | `cmd/gooo/**` |
| `agent/cli-check-current2` | `cmd/gooo/**` |
| `agent/codegen-followup` | `docs/research/codegen-reproducibility.md` |
| `agent/codegen-fixture-adapter` | `docs/research/codegen-fixture-adapter.md` |
| `agent/codegen-hypotheses` | `docs/research/codegen-experiments.md` |
| `agent/codegen-research` | `docs/research/codegen.md` |
| `agent/conformance-fuzz` | `internal/conformance/fuzz/**`, `internal/syntax/**` |
| `agent/dependency-cycle-detector` | `internal/detection/cycles/**` |
| `agent/detection` | `internal/detection/**` |
| `agent/detection-cycles` | `internal/detection/cycles/**` |
| `agent/docs` | `AGENTS.md`, `CONTRIBUTING.md`, `README.md`, `docs/**`, `examples/**` |
| `agent/formatter` | `internal/formatter/**` |
| `agent/freshness-detection` | `internal/detection/freshness/**` |
| `agent/freshness-research` | `internal/research/freshness/**`, `docs/research/**` |
| `agent/fuzz-conformance` | `internal/conformance/fuzz/**` |
| `agent/generator` | `internal/generator/**` |
| `agent/generator-fixtures-current` | `internal/generator/**` |
| `agent/generator-fixtures-current2` | `internal/generator/**` |
| `agent/generator-fixtures-current5` | `internal/generator/**` |
| `agent/generator-fixtures-current7` | `internal/generator/**` |
| `agent/go-version` | `go.mod` (toolchain directives only) |
| `agent/grammar-followup` | `docs/research/grammar.md` |
| `agent/grammar-research` | `docs/research/grammar.md` |
| `agent/grammar-review` | `docs/research/grammar.md` |
| `agent/integration-governance` | `docs/governance/**` |
| `agent/integration-governance-followup` | `docs/governance/integration-promotion.md` |
| `agent/line-cap-detector` | `internal/detection/linecaps/**` |
| `agent/linecaps` | `internal/detection/linecaps/**` |
| `agent/lsp` | `internal/lsp/**` |
| `agent/lsp-contracts` | `docs/research/lsp.md` |
| `agent/lsp-experiments` | `docs/research/lsp.md` |
| `agent/lsp-research` | `docs/research/lsp.md` |
| `agent/performance` | `internal/detection/performance/**` |
| `agent/performance-regression` | `internal/detection/performance/**` |
| `agent/protected-regions` | `internal/detection/protectedregions/**` |
| `agent/prototype-conformance` | `internal/conformance/**` |
| `agent/prototype-detection` | `internal/detection/**` |
| `agent/prototype-formatter` | `internal/formatter/**` |
| `agent/prototype-provenance` | `internal/provenance/**` |
| `agent/prototype-query` | `internal/query/**` |
| `agent/prov-o-research` | `docs/research/prov-o.md` |
| `agent/provenance-evidence` | `internal/provenance/**` |
| `agent/provenance-freshness-detector` | `internal/detection/freshness/**` |
| `agent/provenance-store` | `internal/provenance/**` |
| `agent/query-engine` | `internal/query/**` |
| `agent/query-research` | `docs/research/query.md` |
| `agent/roundtrip-detector` | `internal/detection/roundtrip/**` |
| `agent/roundtrip-detection` | `internal/detection/roundtrip/**` |
| `agent/security` | `docs/research/security.md` |
| `agent/security-research` | `docs/research/security.md` |
| `agent/self-hosting-bootstrap` | `docs/research/self-hosting.md`, `internal/bootstrap/**` |
| `agent/semantic` | `internal/semantic/**` |
| `agent/semantic-delta-detector` | `internal/detection/semanticdelta/**` |
| `agent/semanticdelta` | `internal/detection/semanticdelta/**` |
| `agent/syntax` | `internal/syntax/**` |
| `agent/testing-research` | `docs/research/testing.md` |
| `agent/testing-research-contracts` | `docs/research/testing.md` |
| `agent/testing-research-followup` | `docs/research/testing.md` |
| `agent/transformation-effect-v5` | `.github/agent-scope-table.md`, `.github/ci-governance.json`, `.github/workflows/transformation-effect.yml`, `internal/meta/transformationeffect/**`, `internal/verify/scope_part01.go`, `scripts/transformation-effect/**` |
| `agent/metric-transition-v6` | `.github/agent-scope-table.md`, `.github/ci-governance.json`, `.github/workflows/metric-transition.yml`, `internal/meta/metrictransition/**`, `internal/verify/scope_part01.go`, `scripts/metric-transition/**` |
| `agent/metric-counterfactual-v7` | `.github/agent-scope-table.md`, `.github/ci-governance.json`, `.github/workflows/metric-counterfactual.yml`, `internal/meta/metriccounterfactual/**`, `internal/meta/metriccounterfactualio/**`, `internal/meta/metriccounterfactualverify/**`, `internal/verify/scope_part01.go`, `scripts/metric-counterfactual/**` |
| `agent/metric-intervention-v8` | `.github/agent-scope-table.md`, `.github/ci-governance.json`, `.github/workflows/metric-intervention.yml`, `internal/meta/metricintervention/**`, `internal/meta/metricinterventionverify/**`, `internal/verify/scope_part01.go`, `scripts/metric-intervention/**` |
| `agent/zerolang-experiments` | `docs/research/zerolang.md` |
| `agent/zerolang-research` | `docs/research/zerolang.md` |
