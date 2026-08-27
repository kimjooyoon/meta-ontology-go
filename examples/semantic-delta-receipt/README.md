# Semantic delta receipt

This experiment makes three change layers explicit for a pair of Gooo source
files:

| Layer | What is measured | What it cannot prove alone |
| --- | --- | --- |
  | `textual_delta` | raw bytes, byte digests, and positional byte mismatches | that meaning changed |
| `structural_delta` | canonical nodes and directed semantic facts | that every claim interpretation is valid |
| `semantic_claim_delta` | stable claim IDs, changed claim values, and transitions | behavior outside the bounded Gooo grammar |

The fixed denominator is version `v2` with five cases: `equivalent`,
`semantic-change`, `value-program-change`, `indeterminate`, and
`ambiguous-match`. The first changes text while preserving meaning; the next
two change modeled semantics; the last two remain `FAIL_CLOSED /
LOWER_RESOLUTION` at their exact stage/step/reason coordinates.

The producer records the three layers. The consumer is an independent
adjudicator in a separate package: it rereads both raw sources and recomputes the expected layers
before accepting the receipt. It rejects a tampered receipt and records effects
from path-plus-content snapshots without claiming mutation authority. The
meta-operation is
`separate-text-structural-semantic-deltas`.

The raw decision and semantic decision are separate receipt fields. Claim
transitions persist `OPEN`, `DISCHARGED`, or `REFUTED`; an unknown subject is
reported as `FAIL_CLOSED / LOWER_RESOLUTION` with the exact stage
`bind-subject`, step `observe-checkout-sha`, and reason
`SUBJECT_SHA_UNAVAILABLE`.

Object propositions are not preservation propositions. A changed or removed
before object claim refutes its separate preservation row; an after-only object
claim is a canonical source observation and is discharged. Proposition and
preservation digests stabilize claim IDs, and old rows remain in the receipt.
Object and preservation claim inventories are sets: both sides must be unique
and are compared after canonical sorting; array order is not semantic. The
transition identity digest is version `v2` and is independently rebuilt from
canonical rows `(claim_id, from_status, to_status, stage, step, reason,
target_semantic_digest)`, sorted by claim ID. The target semantic digest is the
claim's after-source semantic digest, falling back to its before-source
semantic digest when no after target exists.
Stable claim identity is version `v3`: proposition kind, canonical semantic
fact target address (`subject\x00predicate\x00object`), and stable relation role; the bounded pair uses canonical before/after semantic addresses. Raw source paths/digests and observed
semantic digests are evidence/provenance fields and cannot recreate a claim.
The semantic projection explicitly models node identity, entity fields,
activity value programs, relation facts, and the IR semantic fingerprint. Its
declared projection component-kind coverage is `5/5 = 10000` basis points, not
whole-language semantic coverage; an uncovered StableHash field is
`UNMODELED_SEMANTIC_COMPONENT_CHANGED`, never preservation.

The v1 expectation artifact is a `HISTORICAL_SCHEMA_MIGRATION` and its
removed/added IDs are not persistence evidence. Actual v3 persistence is
computed from the five baseline/alternate source pairs declared in
`claim-identity-persistence-manifest.json`. Producer and consumer each read
both raw pairs and emit per-claim stable identity, proposition, target, and raw
evidence fields. The fixed result is same-slot persistence: stable identity
31/31, evidence-only changes 31/31, semantic evidence preserved 31/31,
semantic targets preserved 31/31, and
raw-only claim recreation 0/31. Expectation conformance remains a separate
5/5, 31-row check.

The report distinguishes `reconstruction_exact` from
`persistence_satisfied`. Only the latter can become `FIXED_POINT / EXACT`, and
only after unique inventories, changed raw evidence, preserved semantic
evidence and targets, complete evidence-only changes, zero raw-only
recreation, and explicit producer/consumer fixed-point decisions all hold.
Identical raw observations fail closed as `RAW_EVIDENCE_UNCHANGED`; semantic
target drift fails closed as `SEMANTIC_TARGET_CHANGED`; ordinary added/removed
claims fail closed as `CLAIM_SET_CHANGED`. The checked-in
`claim-identity-fault.json` is a separate diagnostic artifact, not a Gooo
source contract: producer and independent consumer packages reconstruct the
real `.gooo` pairs, read the artifact, and emit opaque receipts. The witness
only compares those receipts. Each receipt binds the artifact path, byte count,
digest, rule, old/new inventories, mapping digest, reference counts, dangling
count, raw evidence count, and alpha-equivalent graph digests. Both
implementations must report `FAIL_CLOSED / LOWER_RESOLUTION /
CLAIM_RECREATED_DUE_ONLY_TO_RAW_DIGEST`; stale references, swapped edges, and
duplicate edges have distinct fail-closed reasons. Normal source projection
never reads a marker to alter identity. The consumer import boundary is checked
as forbidden producer imports `0/0`.

## Research decisions

The experiment adopts the following principles from primary sources:

1. Semantic diffing is a separate question from syntactic diffing. Microsoft
   Research describes semantic diffing as comparing program abstractions and
   explaining behavioral differences, rather than only changed lines. See
   [Abstract Semantic Diffing of Evolving Concurrent Programs](https://www.microsoft.com/en-us/research/publication/abstract-semantic-diffing-evolving-concurrent-programs/).
2. Validation is per translation/change, not a claim that every future
   transformation is correct. LLVM's translation-validation description says
   each compiler run is followed by a validation pass for the produced target.
   See [Program Analysis for Compiler Validation](https://llvm.org/pubs/2008-11-PASTE-CompilerValidation.html).
3. Semantic preservation must be stated over observable meaning. CompCert's
   [formal semantic-preservation theorem](https://compcert.org/man/manual001.html)
   is the boundary we echo: a successful translation may be credited only when
   the resulting meaning is related to the source meaning.

We reject raw line-count equality as a semantic proof, text equality as the
definition of equivalence, and approximation as an `EXACT` result. An
unparseable or unsupported source is evidence that the validator is unable to
decide, not evidence that the source is equivalent.

## Meta-operation contract

The checked-in `main.gooo` is executable contract input: it declares the layer
identities, five semantic component kinds, three claim kinds, decision policy,
claim identity and transition identity recipes, ledger recipe, five case input
recipes with before/after addresses, and the `v2:5` denominator. Producer and
consumer parse/lower that source independently;
the Go denominator and JSON expected conclusions are validator expectations.
The suite reports the fixed `5/5 = 10000` contract reproduction. Its subject
semantic equivalence remains separately `NOT_ASSERTED`.

For a known pair:

```text
textual_changed && structural_delta == empty && semantic_claim_delta == empty
  => SEMANTIC_PRESERVED / FIXED_POINT / EXACT
textual_changed && (structural_delta != empty || semantic_claim_delta != empty)
  => SEMANTIC_CHANGED / DELTA_OBSERVED / EXACT
```

Any failed source projection, ambiguous claim match, or receipt replay is
`INDETERMINATE` and `FAIL_CLOSED`; projection failure uses
`project-source/parse-lower`, while ambiguity uses `claim-delta/match-claims`.
An unavailable checkout SHA uses `bind-subject/observe-checkout-sha` with
`SUBJECT_SHA_UNAVAILABLE`; an observed mismatch uses
`REFUTED_SUBJECT_SHA_MISMATCH`. Net repository equality is observed as
`NET_REPOSITORY_STATE_UNCHANGED`, while transient writes and mutation authority
remain `UNKNOWN`. No activation or promotion is implied. The suite's
`FIXED_POINT` is only fixed five-case contract reproduction.

## Falsification

The claim is falsifiable. Reordering declarations/comments should change only
raw evidence digests; changing a stable ID, output entity, relation, or activity value
program must move the receipt to `SEMANTIC_CHANGED`. Multiple claims in one
slot must remain `AMBIGUOUS_CLAIM_MATCH`; unsupported grammar must remain
`SEMANTIC_TRANSLATION_VALIDATION_UNAVAILABLE`. Mutating any receipt layer and
resealing it without recomputing the independent projection must be rejected.
The duplicate-stable-ID fixture is intentionally outside the fixed denominator:
repeated identical semantic facts must produce duplicate IDs and fail closed.
