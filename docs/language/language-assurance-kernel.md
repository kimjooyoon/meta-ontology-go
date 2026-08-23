# Language assurance kernel v1

This kernel measures whether a candidate change can manufacture the evidence that authorizes it. It does not measure language completeness and does not change the saturated `gooo/language-readiness-artifact/v2` denominator.

## Frozen denominator

`gooo/language-assurance-denominator/v1` contains exactly 12 obligations. The first report is intentionally partial: 3 obligations are operating, 9 are not implemented, and implementation coverage is exactly `3 / 12 = 2500 basis points`.

| Priority | Required indicator | Meta operation | Initial status |
| --- | --- | --- | --- |
| P0 | `gooo.metric.governance.self-minting-paths.v1` | `detect-self-minting-paths` | OPERATING |
| P0 | `gooo.metric.governance.role-conflict-paths.v1` | `detect-role-conflict-paths` | OPERATING |
| P0 | `gooo.metric.epistemic.unknown-laundering.v1` | `detect-unknown-laundering` | OPERATING |
| P0 | `gooo.metric.evidence.exact-snapshot-binding.v1` | `bind-exact-snapshot` | NOT_IMPLEMENTED |
| P0 | `gooo.metric.evidence.raw-reconstruction.v1` | `reconstruct-raw-evidence` | NOT_IMPLEMENTED |
| P0 | `gooo.metric.effects.write-set-exactness.v1` | `observe-exact-write-set` | NOT_IMPLEMENTED |
| P1 | `gooo.metric.semantic.source-backed-authority.v1` | `bind-source-backed-authority` | NOT_IMPLEMENTED |
| P1 | `gooo.metric.semantic.candidate-leakage.v1` | `detect-candidate-leakage` | NOT_IMPLEMENTED |
| P1 | `gooo.metric.semantic.changed-surface-receipt-totality.v1` | `totalize-changed-surface-receipts` | NOT_IMPLEMENTED |
| P1 | `gooo.metric.operation.rollback-integrity.v1` | `verify-rollback-integrity` | NOT_IMPLEMENTED |
| P2 | `gooo.metric.capability.vertical-slice-closure.v1` | `close-vertical-slice` | NOT_IMPLEMENTED |
| P2 | `gooo.metric.ecosystem.external-conformance.v1` | `verify-external-conformance` | NOT_IMPLEMENTED |

The report embeds the denominator and its SHA-256 digest. A missing meta operation remains `NOT_IMPLEMENTED`; it is never converted to metric value zero.

## Operating indicators

The evaluator emits exactly five observations:

| Class | Count | Meaning |
| --- | ---: | --- |
| OUTCOME | 1 | frozen-denominator implementation coverage |
| DRIVER | 1 | transaction evidence-group coverage |
| GUARDRAIL | 3 | self-minting, role-conflict, and UNKNOWN-laundering paths |

Each indicator names its producer and meta operation. The proof choices cover the Munchhausen trilemma explicitly: `FOUNDATION`, `COHERENCE`, and `REGRESSION`.

## Exact decision rules

Self-minting is one path for each authority rule whose `authored_by` and `promoted_by` principals are equal.

Role conflict is one path for each principal and each configured incompatible pair:

- `CONTRACT_AUTHOR + EVALUATOR_AUTHOR`
- `IMPLEMENTER + PROMOTER`
- `EVALUATOR_AUTHOR + AUDITOR`
- `POLICY_ADOPTER + PROMOTER`
- `ADAPTER_AUTHOR + AUDITOR`

UNKNOWN laundering is one path for each transition from `UNKNOWN` to `PASS`, `FIXED_POINT`, `AUTHORIZED`, or `ALLOW`. Any top-level `UNKNOWN` produces `FAIL_CLOSED / ASSURANCE_TOP_DECISION_UNKNOWN`, even when its output is `FAIL` or `BLOCK`. Missing evidence produces JSON `null`, lowers resolution to `UNKNOWN`, and never becomes a zero-valued success.

An exact, violation-free transaction receives `ALLOW_LIMITED`, not `PASS`, because only 3 of 12 obligations operate. A detected governance path receives `BLOCK`.

## Executable use cases

| Fixture | Candidate decision | Exact guardrail path |
| --- | --- | ---: |
| `independent.json` | `ALLOW_LIMITED` | 0 |
| `self-minting.json` | `BLOCK` | self-minting 1 |
| `role-conflict.json` | `BLOCK` | role-conflict 1 |
| `unknown-laundering.json` | `FAIL_CLOSED` | UNKNOWN-laundering 1 |

The witness writes only to a path outside the repository. CI evaluates every fixture twice, compares report bytes, checks the exact counts, and uploads the first report set.

## Go 1.27 and metaprogramming boundary

CI uses Go `1.27.0` and `go fix -diff`. The official Go command contract makes `-diff` non-mutating and non-zero when a modernization patch exists, so CI can enforce fixes without writing to the candidate tree. Go 1.27 adds the `atomictypes`, `embedlit`, `slicesbackward`, and `unsafefuncs` modernizers. See the [Go 1.27 release notes](https://go.dev/doc/go1.27) and [`go fix` command documentation](https://go.dev/cmd/go/).

[gomacro](https://github.com/cosmos72/gomacro) is a useful comparison point for runtime Go interpretation and AST-to-AST Lisp-like macros. This PR does not claim those capabilities. Its meta-level contribution is narrower: executable code observes the authority graph of another code-changing transaction and emits a replayable, fail-closed decision.
