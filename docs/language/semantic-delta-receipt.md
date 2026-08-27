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

Object claim type IDs are proposition digests, while persistent claim IDs use
claim kind, the canonical semantic fact target address
`subject\x00predicate\x00object`, and stable relation role. The bounded pair
target is the canonical `before-semantic-address->after-semantic-address`; raw
paths remain evidence only. Raw
source paths/digests and observed semantic digests are evidence fields, never
identity inputs. Preservation IDs bind the before proposition ID, canonical
target address, and preservation role. Each transition
records proposition/evidence/previous-event/transition digests in an append-only
chain. The semantic-change case therefore refutes the old
payment preservation proposition and discharges the new reversal observation;
it does not claim that the reversal proposition is false.

This identity recipe is version `v3`; observation-side roles (`before` and
`after`) are explicit stable relation roles, so the fixed ledger can retain
both source observations without making raw evidence part of identity.

Every ledger row has exactly one transition linked by claim ID and proposition
digest, including every `OPEN`, `DISCHARGED`, and `REFUTED` row. The reported
claim-status coverage is `claims_with_explained_status / total_claims`; it is
not padded by duplicate events. A comment-only intervention changes the raw
digest while preserving semantic digests, the semantic decision, and the exact
logical transition sequence (`kind`, status endpoints, preservation target,
stage, step, and reason). Evidence digests may still change because they
deliberately bind the observed raw source, while persistent IDs remain unchanged.
The evolution receipt reports historical v1-to-v3 migration separately from
actual persistence. `historical_migration_removed` and
`historical_migration_added` describe the old evidence-bound inventory being
replaced; they are never persistence numerators. Persistence is computed from
two real v3 observations for each of the five fixed source pairs, using the
checked-in `claim-identity-persistence-manifest.json`. Producer and consumer
each read both files and independently rebuild the mapping. The resulting
`stable_identity_preserved=31/31`, `evidence_only_changes=31/31`,
`semantic_evidence_preserved_on_nonsemantic=31/31`,
`semantic_target_preserved_on_nonsemantic=31/31`, and
`claim_recreated_due_only_to_raw_digest=0/31` mean the same observation slot
across two raw interventions, not before-to-after preservation within one
pair. Each report row includes both implementations' source-pair
observations and all per-claim stable/evidence fields.

`reconstruction_exact` is intentionally weaker than
`persistence_satisfied`: it only says that the two raw-source reconstructions
and their record mappings agree. Persistence is promoted to `FIXED_POINT` only
when every fixed predicate also holds: unique nonempty inventories, no added or
removed IDs, complete stable identity, changed raw evidence, preserved
semantic evidence and target, evidence-only change for every claim, no
raw-only recreation, and explicit `FIXED_POINT / EXACT` decisions from both
implementations. Identical raw pairs therefore fail closed with
`RAW_EVIDENCE_UNCHANGED`; semantic-target drift fails closed with
`SEMANTIC_TARGET_CHANGED`; ordinary added/removed claims remain
`CLAIM_SET_CHANGED`. A separate identity-fault artifact can deliberately
rekey alternate `StableID` values after both raw-source reconstructions and
rewrite all internal identity references through the same mapping. The artifact
path, byte count, raw digest, and exact mutation rule are bound in
the probe receipt. When that fault leaves semantic evidence and targets equal,
both implementations must classify it as
`FAIL_CLOSED / LOWER_RESOLUTION / CLAIM_RECREATED_DUE_ONLY_TO_RAW_DIGEST`.
This diagnostic fault is outside the canonical producer/consumer source
projection and cannot make a normal Gooo source marker alter identity.
Historical v1-to-v3 migration remains separate accounting and cannot satisfy
these predicates.

The diagnostic rekey is graph-closed: the producer package and independent
consumer package each record a canonical old-to-new StableID bijection, rewrite
`PreservationOf` and every other internal identity reference through that
bijection, and emit an opaque receipt. The producer declares
`forward-map-reverse-normalize/v1`; the consumer declares
`canonical-ordinal-edge-join/v1`. Each binds its algorithm source path, byte
count, and source digest in the receipt; the path is repository-relative and
CI re-reads the two checked-in files to independently recalculate both byte
counts and digests. The witness compares only the
implementation-produced common evidence wire; it does not perform graph
adjudication or feed either receipt into the other implementation. The
missing or invalid binding is immediately
`FAIL_CLOSED / LOWER_RESOLUTION / identity-fault / bind-algorithm-source /
ALGORITHM_SOURCE_UNAVAILABLE`.
consumer constructs a semantic-ordinal inventory, checks injectivity by sorted
adjacent IDs, joins reference edges by ordinal, and hashes an ordinal graph
encoding. Semantic-slot uniqueness uses the fixed `semantic_slot_denominator=7`
for this fixture and is explicit as `semantic_slot_unique / semantic_slot_total`
(`7/7` on the valid input); a duplicate slot is
`IDENTITY_SEMANTIC_SLOT_AMBIGUOUS`. Each receipt records both inventories, mapping digest, reference
denominator, rewritten-reference count, dangling count, raw evidence count, and
alpha-equivalent semantic-graph digests. A stale reference is
`FAIL_CLOSED / LOWER_RESOLUTION / IDENTITY_REFERENCE_CLOSURE_BROKEN`; swapped
mapping edges are `IDENTITY_FAULT_MAPPING_RULE_MISMATCH`; duplicate edges are
`IDENTITY_FAULT_MAPPING_DUPLICATE_EDGE`; an invalid consumer ordinal edge is
`IDENTITY_FAULT_ORDINAL_EDGE_MISMATCH`. Raw-only recreation is admissible only
after both independent graph proofs pass. CI records forbidden imports as
`forbidden_imports_observed=0`, `forbidden_imports_allowed_max=0`, and
`forbidden_import_conformance=1/1`; an observed import is
`FAIL_CLOSED / LOWER_RESOLUTION / PRODUCER_IMPORT_FORBIDDEN`.
The asymmetric probes tamper the producer evidence bytes while comparing them
with a fresh consumer raw reconstruction, then reverse the direction; both
must be mismatches (`2/2`) without passing either receipt into the other
algorithm.

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
runtime inventory, the CI-generated
`claim-transition-expectation-evolution.json` records old/new artifact paths,
bytes, digests, every case's migration added/removed IDs, independently
reconstructed expectation-conformance rows, and the unchanged denominator.
The same receipt separately records the two-observation v3 persistence rows;
the checked-in expectation is never silently overwritten.

The semantic claim-delta manifest is executable evidence: CI passes its raw
bytes to producer-side and independent consumer-side source readers, then
compares both reconstructed added/removed claim sets with the manifest's
validator expectations. A separate persistence probe exercises identical raw
pairs and semantic-target drift so reconstruction agreement cannot be mistaken
for persistence.

The conformance suite's `FIXED_POINT` decision means only that the fixed
five-case contract was reproduced. `subject_semantic_equivalence` is recorded
separately as `NOT_ASSERTED`.

## Falsifiability

The result is falsifiable. Reordering declarations or adding comments should
keep the first case semantically preserved. Changing a stable ID, output entity,
or supported relation must produce a non-empty structural or claim delta.
Putting an unsupported declaration into the source must remain indeterminate.
The fixture `duplicate-stable-id-before.gooo` deliberately repeats one
semantic fact; its duplicate stable identity must remain a fail-closed input,
not a fixed-case success.
Mutating one field of a receipt without resealing it must be rejected by the
independent judge. These tests are intentionally narrower than behavioral
equivalence: the bounded projector does not cover runtime values, macros,
external dependencies, or arbitrary future Gooo syntax.
