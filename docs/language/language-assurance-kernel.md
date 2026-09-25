# Language assurance kernel v1

This kernel measures whether a candidate change can manufacture the evidence that authorizes it. It does not measure language completeness and does not change the saturated `gooo/language-readiness-artifact/v2` denominator.

## Frozen denominator

`gooo/language-assurance-denominator/v1` contains exactly 12 obligations. The current report is intentionally partial: 7 obligations are operating, 5 are not implemented, and integer implementation coverage is exactly `(7 * 10000) / 12 = 5833 basis points`.

| Priority | Required indicator | Meta operation | Initial status |
| --- | --- | --- | --- |
| P0 | `gooo.metric.governance.self-minting-paths.v1` | `detect-self-minting-paths` | OPERATING |
| P0 | `gooo.metric.governance.role-conflict-paths.v1` | `detect-role-conflict-paths` | OPERATING |
| P0 | `gooo.metric.epistemic.unknown-laundering.v1` | `detect-unknown-laundering` | OPERATING |
| P0 | `gooo.metric.evidence.exact-snapshot-binding.v1` | `bind-exact-snapshot` | OPERATING |
| P0 | `gooo.metric.evidence.raw-reconstruction.v1` | `reconstruct-raw-evidence` | OPERATING |
| P0 | `gooo.metric.effects.write-set-exactness.v1` | `observe-exact-write-set` | OPERATING |
| P1 | `gooo.metric.semantic.source-backed-authority.v1` | `bind-source-backed-authority` | OPERATING |
| P1 | `gooo.metric.semantic.candidate-leakage.v1` | `detect-candidate-leakage` | NOT_IMPLEMENTED |
| P1 | `gooo.metric.semantic.changed-surface-receipt-totality.v1` | `totalize-changed-surface-receipts` | NOT_IMPLEMENTED |
| P1 | `gooo.metric.operation.rollback-integrity.v1` | `verify-rollback-integrity` | NOT_IMPLEMENTED |
| P2 | `gooo.metric.capability.vertical-slice-closure.v1` | `close-vertical-slice` | NOT_IMPLEMENTED |
| P2 | `gooo.metric.ecosystem.external-conformance.v1` | `verify-external-conformance` | NOT_IMPLEMENTED |

The report embeds the denominator and its SHA-256 digest. A missing meta operation remains `NOT_IMPLEMENTED`; it is never converted to metric value zero.

## Operating indicators

The evaluator emits exactly eight observations:

| Class | Count | Meaning |
| --- | ---: | --- |
| OUTCOME | 1 | frozen-denominator implementation coverage |
| DRIVER | 3 | transaction evidence-group coverage, exact snapshot binding, and raw reconstruction |
| GUARDRAIL | 4 | self-minting, role-conflict, UNKNOWN-laundering, and exact write-set paths |

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

Exact snapshot binding requires exactly three unique bindings named `authority_routes`, `role_bindings`, and `decision_transitions`. Its value is `(bindings matching the evaluated subject SHA * 10000) / 3`. A missing binding produces JSON `null` and `FAIL_CLOSED / ASSURANCE_EVIDENCE_UNKNOWN`; a known mismatch produces `BLOCK / ASSURANCE_SNAPSHOT_BINDING_MISMATCH`. One mismatch is exactly `2 / 3 = 6666 basis points`.

Raw reconstruction requires exactly one receipt from `gooo-independent-json-reconstructor-v1`. The separate standard-library command does not import the subject evaluator: it reconstructs normalized observations and the candidate decision from the exact-subject raw transaction, pins the frozen denominator digest, and emits a receipt. Missing receipt evidence produces JSON `null` and `FAIL_CLOSED / ASSURANCE_EVIDENCE_UNKNOWN`. A structurally valid receipt that differs from the evaluator's expected normalized observation produces exactly `0 / 1 = 0 basis points` and `BLOCK / ASSURANCE_RAW_RECONSTRUCTION_MISMATCH`.

An exact, violation-free transaction receives `ALLOW_LIMITED`, not `PASS`, because only 7 of 12 obligations operate. A detected governance, exact-snapshot, or raw-reconstruction violation receives `BLOCK`.

## Executable use cases

| Fixture | Candidate decision | Exact guardrail path |
| --- | --- | ---: |
| `independent.json` | `ALLOW_LIMITED` | 0 |
| `self-minting.json` | `BLOCK` | self-minting 1 |
| `role-conflict.json` | `BLOCK` | role-conflict 1 |
| `unknown-laundering.json` | `FAIL_CLOSED` | UNKNOWN-laundering 1 |
| `snapshot-mismatch.json` | `BLOCK` | snapshot mismatch 1; exact binding 6666 bps |
| `raw-reconstruction-mismatch.json` | `BLOCK` | reconstruction mismatch 1; reconstruction 0 bps |

The committed fixtures are transaction templates. CI binds their three evidence groups to the exact checked-out head, runs the evaluator-independent reconstructor twice, compares receipt bytes, and injects the receipt before evaluation. The two mismatch fixtures alter one snapshot binding or one reconstructed candidate after receipt generation. Both commands write only outside the repository. CI evaluates every materialized fixture twice, compares report bytes, checks exact counts, and uploads raw transactions, reconstruction receipts, reports, and their manifest together.

## Go 1.27 and metaprogramming boundary

CI uses Go `1.27.0` and `go fix -diff`. The official Go command contract makes `-diff` non-mutating and non-zero when a modernization patch exists, so CI can enforce fixes without writing to the candidate tree. Go 1.27 adds the `atomictypes`, `embedlit`, `slicesbackward`, and `unsafefuncs` modernizers. See the [Go 1.27 release notes](https://go.dev/doc/go1.27) and [`go fix` command documentation](https://go.dev/cmd/go/).

[gomacro](https://github.com/cosmos72/gomacro) is a useful comparison point for runtime Go interpretation and AST-to-AST Lisp-like macros. This PR does not claim those capabilities. Its meta-level contribution is narrower: one executable observes an authority graph, while a separately implemented executable reconstructs the same normalized conclusion from raw evidence; disagreement remains visible and blocks promotion.


## Exact write-set meta-operation

`gooo.metric.effects.write-set-exactness.v1` is operated by
`observe-exact-write-set / ObserveExactWriteSet / REGRESSION`. An independent
standard-library observer snapshots files before and after execution, derives the
changed path set, and compares it with the declared path set. Equal sets produce
`10000 bps / PASS / EXACT`; unequal sets produce `0 bps / BLOCK / EXACT`;
unavailable evidence produces
`FAIL_CLOSED / WRITE_SET_EVIDENCE_UNKNOWN / INVARIANT_ONLY`.

The receipt binds the exact subject SHA and frozen denominator digest. It is
produced in CI while the evaluator keeps repository writes at exactly zero.
