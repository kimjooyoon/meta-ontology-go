# SplitGo operation conformance v1

## Decision boundary

This document fixes the denominator for the SplitGo operation before an
evaluator, runtime adapter, or CI promotion exists.

- Contract ID: gooo/source-splitter/split-go-conformance/v1
- Schema: gooo/operation-conformance-contract/v1
- Denominator: 6 indicators
- Oracle corpus: 18 cases, exactly one PASS, one FAIL, and one UNKNOWN
  case per indicator
- Unknown policy: PRESERVE_UNKNOWN_AND_BLOCK
- Passing operation decision: 6 PASS, 0 FAIL, 0 UNKNOWN
- Blocking operation decision: any FAIL or any UNKNOWN
- Assurance grade: E0_SAME_REPOSITORY_CONTRACT
- Readiness transition: none

The machine-readable authority is
examples/source-splitter-conformance/contract.json. The Gooo meta-program in
examples/source-splitter-conformance/main.gooo connects the fixed indicators
to operation evidence and a fail-closed decision. Neither artifact evaluates
the implementation.

## Fixed indicator registry

| ID | Kind | Munchhausen route | Exact subject |
| --- | --- | --- | --- |
| filesystem.atomic-replacement/v1 | guardrail | coherence | replacement mode and writes outside declared targets |
| go.filename.build-semantics/v1 | driver | foundation | selected Go file set under one pinned build context |
| go.header.preserved/v1 | guardrail | regression | byte digest of the pre-package header |
| go.import.identity/v1 | outcome | foundation | multiset digest of import alias and path tuples |
| go.initialization.order/v1 | outcome | foundation | initialization graph and lexical-order digests |
| go.package.conformance/v1 | outcome | foundation | parsed selected files and their package names |

The role denominator is outcome 3/6, driver 1/6, guardrail 2/6. The
Munchhausen route denominator is foundation 4/6, coherence 1/6, regression
1/6. These counts are registry facts, not quality scores.

## Deterministic resolution

Each oracle observation carries an explicit evidence_complete value.

- evidence_complete=false resolves to UNKNOWN.
- Complete evidence satisfying the named rule resolves to PASS.
- Complete evidence violating the named rule resolves to FAIL.
- No unrecognized value resolves to PASS or FIXED_POINT.
- UNKNOWN is preserved in evidence and makes the operation decision BLOCK.

The contract measures write effects through
writes_outside_declared_targets. The current stage only defines its oracle
shape. Runtime write observation remains 0/1 until the adapter and CI stages
produce exact-head evidence.

## Authority separation

The pre-registered authorities are intentionally disjoint.

| Stage | Owned artifact | Current completion |
| --- | --- | --- |
| Contract | this document, Gooo meta-program, oracle JSON | indicators 6/6; oracle declarations 18/18 |
| Evaluator | internal/meta/operationconformance/**, scripts/source-splitter/** | rules 0/6 |
| Adapter | generation registry and transformation-effect binding | bindings 0/1 |
| CI | workflow and exact-head conformance collector | enforcement 0/1 |

This contract PR cannot edit evaluator, adapter, CI, or governance paths.
Consequently it cannot widen its own authority or promote its own result. A
later evaluator must consume this same contract ID, denominator version, and
oracle cases. A later adapter must preserve UNKNOWN, and CI must bind evidence
to the exact PR head.

## Claim budget

The following are exact facts for this stage:

- Indicator definitions: 6/6
- Oracle declarations: 18/18
- Evaluated indicators: 0/6
- Runtime-observed indicators: 0/6
- Independent implementations: 0
- Adapter bindings: 0/1
- CI enforcement bindings: 0/1
- Readiness transitions: 0
- Language-completion claims: 0

Therefore this stage defines what future success means; it does not claim that
SplitGo succeeds. Repository Go/Gooo line counts and directory topology
counts remain separate observational metrics and cannot substitute for these
operation-semantic indicators. The project-root README exception is unchanged
and is outside this file-operation denominator.

## Primary foundations

- Go 1.27 release notes: https://go.dev/doc/go1.27
- Go language specification: https://go.dev/ref/spec
- go/build.Context.MatchFile: https://pkg.go.dev/go/build#Context.MatchFile
- gomacro AST macro model: https://github.com/cosmos72/gomacro

The gomacro reference motivates explicit program-to-program transformation
boundaries. Its unrestricted runtime side effects are not adopted here.
