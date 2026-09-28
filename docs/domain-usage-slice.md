# Domain usage slice

This document turns the domain-completeness contract into user-facing slices that can
be measured without treating a score as a language-completeness claim.

## Operating rule

A slice is a bounded user operation with an explicit input, an observable output,
and a provenance tuple. The tuple is the exact combination of source, declaration,
catalog, schema, evidence, and toolchain identities. A cache hit is never semantic
or test evidence.

A slice may be `AVAILABLE`, `DEFERRED`, `UNKNOWN`, or `OUT_OF_SCOPE`. The state is
part of the receipt and must not be inferred from a high aggregate score.

## Initial gooo slices

| Slice | User operation | Required evidence | Safe boundary |
| --- | --- | --- | --- |
| `declaration-to-ir` | inspect a declaration and derive its intermediate representation | source digest, declaration digest, IR digest, schema digest | generation is not execution |
| `ir-to-canonical-source` | produce a canonical declaration from the representation | IR digest, generated-source digest, schema digest | generated text is not semantic proof |
| `canonical-round-trip` | reparse generated source and compare representations | source digest, generated digest, reverse-observation digest | mismatch remains visible |
| `capability-discovery` | ask what gooo can expose for a declaration | query digest, declaration digest, catalog digest, evidence digest | discovery never invokes a provider |
| `lsp-next-operation` | show a hover, guide, or editor action for the next safe inspection | LSP evidence digest, first missing stage, next non-executing operation | an editor action is not authorization |
| `security-boundary-observation` | describe a workload-identity or network boundary that is still external | JEV observation digest, boundary evidence, schema digest | no credential issuance or execution |
| `feedback-window` | compare bounded feedback observations across a declared window | window digest, metric numerator and denominator, provenance continuity | a trend is not a correctness oracle |

## Metric contract

For each slice, record the numerator, denominator, excluded identifiers, and first
missing stage.

- Scope coverage is `in_scope_items / declared_scope_items`.
- Available coverage is `available_items / eligible_in_scope_items`.
- Evidence completeness is `complete_evidence_items / eligible_items`.
- Provenance continuity is `tuple_valid_items / observed_items`.
- Domain utility is `useful_outcomes / attempted_operations` where usefulness has a
  predeclared user-facing criterion.
- Investment efficiency is `utility_delta / bounded_cost_units`.

A denominator of zero is `UNKNOWN`, not one hundred percent. An excluded item must
carry an identifier and reason. A metric without its exact source and schema tuple
is not comparable across revisions.

## Investment decision

Before implementing a candidate slice, record:

1. the user operation and why it matters;
2. the current unknown identifiers and first missing stage;
3. the bounded implementation and CI cost estimate;
4. the metric expected to change and its minimum meaningful delta;
5. a falsifier that would reject the hypothesis;
6. a rollback condition that prevents a misleading score increase.

Proceed only when the candidate adds a user-visible operation and a durable evidence
path. Defer when the operation is useful but its boundary or denominator is not
defined. Reject when it only increases a score, duplicates an existing receipt, or
requires unverified authority.

## Feedback loop

The workflow is:

1. observe a user operation and persist its exact receipt;
2. reverse-observe the source, generated artifact, and declared boundary;
3. classify the slice state without collapsing `DEFERRED` or `UNKNOWN`;
4. compare the bounded metric delta with its prior exact tuple;
5. select the next investment or record a falsifier;
6. preserve the receipt even when the metric later decreases.

This makes progress reversible and auditable. It does not claim that the language can
compress arbitrary implementation knowledge or replace domain-specific semantics.
