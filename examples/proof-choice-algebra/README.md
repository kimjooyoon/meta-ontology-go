# Semantic proof-choice algebra

This experiment treats the justification route as a meta value derived from
semantic evidence. It does not copy Lean, Rocq, or another language type
system. The input is ordinary Gooo: each `activity ... computes "..."` value
program lowers into a canonical semantic value. Comments are never parsed as
evidence.

## Input and route selection

Each semantic value may be an `observation`, `claim`, or `metric`.

- A claim carries `prior_state`, `dependencies`, observations, and an
  `admissible_routes` set.
- An observation carries a provenance, evidence kind, predicate, value, and
  observed bit.
- A metric carries observation references and exactly three observed slots.

The route is computed from observation signals: stable identity selects
`FOUNDATION`, agreeing relations select `COHERENCE`, and an equal replay
selects `REGRESSION`. Exactly one matching admissible route is `EXACT`; none
is `LOWER_RESOLUTION` or `UNKNOWN` when evidence is missing; more than one is
`FAIL_CLOSED`. The source cannot contain a final `choice`, `numerator`, or
`denominator` field.

The algebraic composition operator is semantic-value union:

```text
x ⊕ y = x ∪ y             when IDs are disjoint
x ⊕ x = x                 for an exact duplicate value
x ⊕ y = FAIL_CLOSED       when one ID has conflicting semantic values
```

The fixed denominator is calculated as `len(metric.slots)` and the numerator
as the number of slots whose `observed` bit is true. No denominator is supplied
by a producer.

## Ledger and receipts

An `OPEN` prior claim becomes `DISCHARGED` only after an exact route, remains
`OPEN` for lower or unknown resolution, and becomes `REFUTED` for an explicit
route contradiction. Ledger entries are append-only and carry stage, step,
reason, evidence digest, route resolution, and observation provenance.

The producer receipt contains both the raw source digest and canonical
semantic digest, source reconstruction counts, derived choices, metric slot
counts, the ledger, and CI before/after effect observations. The separate
consumer parses and lowers the raw source itself, reconstructs the values, and
checks both digests, the receipt digest, the ledger, and effect snapshots. A
receipt-only call reports independent evidence `0` and cannot certify a pass.

## Official proof-tool principles

The [Rocq Reference Manual](https://rocq-prover.org/doc/master/refman/index.html)
describes a small kernel that finally verifies tactic-produced proof output.
We adopt the producer/checker boundary, but reject treating producer output as
self-authenticating.

The [Lean Language Reference](https://lean-lang.org/doc/reference/latest/)
describes a minimal kernel checking proof terms produced by tactics. We adopt
independent reconstruction and digest binding, but reject Lean proof terms,
elaboration, and dependent types because this is an evidence-route algebra.

## Fixtures and falsification

`foundation.gooo`, `coherence.gooo`, and `regression.gooo` each derive one route;
`combined.gooo` composes all three. `lower-resolution.gooo` has evidence that
does not match its admissible route. `unknown.gooo` references absent evidence.
`contradiction.gooo` makes two admissible routes match and must fail closed.

CI also creates two interventions. Changing only the comment must change the
raw digest while preserving the semantic digest and decision. Changing the
`computes` value must change the semantic digest, route resolution, and claim
transition. These probes falsify the design if either distinction disappears.
