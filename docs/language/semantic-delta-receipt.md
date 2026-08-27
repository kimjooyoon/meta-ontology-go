# Semantic delta receipt

This independent philosophy experiment treats a source change as three
different objects:

1. `textual_delta`: the raw byte change between two `.gooo` files;
2. `structural_delta`: the change in normalized nodes and directed facts; and
3. `semantic_claim_delta`: the change in stable claims, including explicit
   claim transitions.

It is a Gooo meta-operation, not a second line-oriented diff. The operation is
`separate-text-structural-semantic-deltas`. `semanticdeltareceipt.ProduceFiles`
is the producer; `semanticdeltareceiptconsumer.AdjudicateFiles` is a separate
consumer and independent judge. Neither path writes the repository.

## Research and adopted principles

The design is grounded in three primary sources:

- Microsoft Research's [Abstract Semantic Diffing of Evolving Concurrent
  Programs](https://www.microsoft.com/en-us/research/publication/abstract-semantic-diffing-evolving-concurrent-programs/)
  treats semantic diffing as comparison of program abstractions and behavioral
  differences, not just changed lines.
- LLVM's [Program Analysis for Compiler
  Validation](https://llvm.org/pubs/2008-11-PASTE-CompilerValidation.html)
  describes translation validation as a validation pass after each compiler
  run, proving the produced target is a correct translation of that source.
- The CompCert [semantic-preservation
  theorem](https://compcert.org/man/manual001.html) states the desired boundary
  over observable source and target meaning, while allowing compilation to
  fail rather than inventing a result.

The experiment adopts these rules:

| Principle | Implementation |
| --- | --- |
| Syntax and semantics are different observations | byte digests never decide semantic class |
| Validation is per change pair | each receipt is rebuilt and adjudicated from its two raw sources |
| Meaning is anchored to stable identities | nodes and claims use immutable IDs, not display order |
| Approximation is not exact proof | unsupported syntax becomes `INDETERMINATE / FAIL_CLOSED / LOWER_RESOLUTION` |
| Effects are observed, not inferred | CI derives tracked-plus-untracked path-plus-content snapshots; net equality is `NET_REPOSITORY_STATE_UNCHANGED`, while transient writes and mutation authority remain `UNKNOWN` |

The raw decision is recorded independently as `RAW_CHANGED` or
`RAW_FIXED_POINT`; the semantic decision is separately recorded as
`SEMANTIC_PRESERVED`, `SEMANTIC_CHANGED`, or `SEMANTIC_UNKNOWN`. An unknown
subject is `FAIL_CLOSED / LOWER_RESOLUTION` at stage `bind-subject`, step
`observe-checkout-sha`, with reason `SUBJECT_SHA_UNAVAILABLE`; invalid SHA uses
step `validate-sha` and reason `SUBJECT_SHA_INVALID`; an observed mismatch is
`REFUTED_SUBJECT_SHA_MISMATCH`. A parse or
lowering failure is likewise `FAIL_CLOSED / LOWER_RESOLUTION` at stage
`project-source`, step `parse-lower`, with reason
`SEMANTIC_TRANSLATION_VALIDATION_UNAVAILABLE`; there is no separate
`resolution=UNKNOWN` state.

The experiment rejects line-count equality as a semantic proof, raw text equality
as equivalence, and a semantic-diff approximation promoted to `EXACT`. It also
does not claim whole-program behavioral equivalence or compiler correctness.
The `declared_projection_component_kind_coverage_bps` value is only coverage of
the five component kinds declared by `main.gooo`, not whole-language semantic
coverage. Because the `ir-semantic-fingerprint` component is a catch-all
StableHash proposition, an unmodeled semantic branch can be hidden behind that
fingerprint; the receipt therefore records `semantic_equivalence_claim:
NOT_CLAIMED` and does not promote that digest to exact equivalence.

## Fixed denominator and cases

The denominator is fixed at version `v2` with five cases and is never widened
by the evaluator:

| Case | Text | Structure | Claims | Result |
| --- | --- | --- | --- | --- |
| `equivalent` | changed | empty | empty | `SEMANTIC_PRESERVED / FIXED_POINT / EXACT` |
| `semantic-change` | changed | one node plus one relation replacement | one changed claim | `SEMANTIC_CHANGED / DELTA_OBSERVED / EXACT` |
| `value-program-change` | changed | topology unchanged | value-program component changed | `SEMANTIC_CHANGED / DELTA_OBSERVED / EXACT` |
| `indeterminate` | changed | unavailable | unavailable | `INDETERMINATE / FAIL_CLOSED / LOWER_RESOLUTION` |
| `ambiguous-match` | changed | known | multiple candidates | `INDETERMINATE / FAIL_CLOSED / LOWER_RESOLUTION` |

The suite reports `5/5 = 10000` basis points only when all five defined cases
are classified exactly. This is the fixed-contract replay denominator, not a
whole-language semantic-coverage score. The declared projection component-kind
denominator is `5/5`; coverage below `10000` cannot claim exact semantic
equivalence. Repository effects are attested from tracked and untracked
path-plus-content digests before the final receipt is created, with output
artifacts kept under `RUNNER_TEMP`.

## Decision rule

For a pair that both bounded projections can parse:

```text
textual_delta.changed && structural_delta == empty && semantic_claim_delta == empty
  => SEMANTIC_PRESERVED / FIXED_POINT / EXACT only when declared projection component-kind coverage is 10000

textual_delta.changed && (structural_delta != empty || semantic_claim_delta != empty)
  => SEMANTIC_CHANGED / DELTA_OBSERVED / EXACT
```

If either projection cannot parse, the result is `INDETERMINATE` and
`FAIL_CLOSED / LOWER_RESOLUTION` at the exact unavailable coordinates above. A receipt is accepted only if the independent
adjudicator recomputes the text, structure, claims, transitions, digests, and
classification from raw bytes. A digest or field mismatch is
`INVARIANT_ONLY / FAIL_CLOSED`.

## Proof choices and claim transitions

Every indicator carries its `producer`, `consumer`, `meta_operation`, `stage`,
`step`, `reason`, and `proof_choice`:

- `FOUNDATION` binds raw bytes, stable identities, and the canonical graph;
- `COHERENCE` checks that the graph, claim delta, transition, and independent
  verdict agree; and
- `REGRESSION` checks the observed net-state boundary while leaving transient
  writes and mutation authority unknown.

`main.gooo` is the semantic contract source. It declares layer identities,
modeled component kinds (node, entity-field, activity-value-program,
relation-fact, and IR fingerprint), claim kinds, fail-closed policy, append-only
ledger recipe, five case recipes, and denominator `v2:5`. Both producer and
consumer parse/lower it independently, reconstruct the case IDs and target
addresses, and bind its digest into the receipt. The JSON denominator must
match those source-derived recipes exactly; expected conclusions remain
validator expectations rather than source declarations.

A transition records the claim ID, claim kind, status before and after, object
before and after, stage, step, reason, and (for preservation rows) the
preserved object claim ID. Object propositions and cross-version preservation
propositions are separate persistent rows. Persistent statuses are `OPEN`,
`DISCHARGED`, and `REFUTED`:

- bounded semantic equivalence is `OPEN -> DISCHARGED` when preserved,
  `OPEN -> REFUTED` when changed, and `OPEN -> OPEN` when indeterminate;
- every before-object preservation row is discharged when the same proposition
  remains and refuted when it changes or disappears;
- after-only object propositions are source observations and are discharged by
  the canonical observation; they are never refuted merely because they are
  new.

Object claim type IDs are proposition digests, while observation instance IDs
also bind target address, raw source digest, and semantic digest. Preservation
IDs bind the before instance, after instance, and pair evidence. Each transition
records proposition/evidence/previous-event/transition digests in an append-only
chain. The semantic-change case therefore refutes the old
payment preservation proposition and discharges the new reversal observation;
it does not claim that the reversal proposition is false.

Every ledger row has exactly one transition linked by claim ID and proposition
digest, including every `OPEN`, `DISCHARGED`, and `REFUTED` row. The reported
claim-status coverage is `claims_with_explained_status / total_claims`; it is
not padded by duplicate events. A comment-only intervention changes the raw
digest while preserving semantic digests, the semantic decision, and the exact
logical transition sequence (`kind`, status endpoints, preservation target,
stage, step, and reason). Instance/evidence digests may still change because
they deliberately bind the observed raw source.

The receipt also exposes the sorted `claim_id_inventory` and a versioned
`claim_transition_identity_digest`. Version `v2` is the digest of canonical
rows `(claim_id, from_status, to_status, stage, step, reason,
target_semantic_digest)` sorted by claim ID; the target digest is taken from the
after-source claim, falling back to the before-source claim. The independent
consumer reconstructs both fields from the raw pair, so a replaced claim with
the same transition count cannot pass. Inventory order is not semantic: both
sides must be unique before canonical sorting. The fixed five-case suite also
compares the independent reconstruction with the checked-in validator
expectation artifact at
`examples/semantic-delta-receipt/claim-transition-expectations.json`; the
producer never reads that artifact. The comparison records the artifact raw
digest, a digest of the exact case row, and before/after raw and semantic source
addresses. Its fixed claim-count contract is `7,7,7,3,7` with total `31`; a
different count requires an explicit denominator-evolution receipt. Missing or
invalid expectation data, missing source sides, and an unavailable raw-pair
reconstruction each retain their own `stage`, `step`, and `reason` at
`LOWER_RESOLUTION`. The CI artifact generated by
`--tamper-matrix` carries a separate denominator ID, exact expected and
observed tamper IDs, four exact replay context IDs, fixed totals, rejection
counts, and basis-point coverage. The fixed tamper inventory is 12 IDs and the
replay-context inventory is 4 IDs; duplicate, missing, extra, or substituted
IDs fail closed. When the fixed expectation is reconciled to a source-derived
runtime inventory, `claim-transition-expectation-evolution.json` records both
artifact digests, every case's added/removed IDs, the unchanged proposition and
target binding, and the unchanged denominator; it is not an implicit overwrite.

The conformance suite's `FIXED_POINT` decision means only that the fixed
five-case contract was reproduced. `subject_semantic_equivalence` is recorded
separately as `NOT_ASSERTED`.

## Falsifiability

The result is falsifiable. Reordering declarations or adding comments should
keep the first case semantically preserved. Changing a stable ID, output entity,
or supported relation must produce a non-empty structural or claim delta.
Putting an unsupported declaration into the source must remain indeterminate.
Mutating one field of a receipt without resealing it must be rejected by the
independent judge. These tests are intentionally narrower than behavioral
equivalence: the bounded projector does not cover runtime values, macros,
external dependencies, or arbitrary future Gooo syntax.
