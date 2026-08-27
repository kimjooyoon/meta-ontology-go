# Claim lifecycle calculus research note

## Question

Can a Gooo meta-operation preserve an assertion while evidence changes its
state, including an explicit refutation, without treating missing evidence as
false or silently dropping the assertion?

The experiment is deliberately smaller than a language feature. It observes
the existing Gooo source relation `activity(inputs) -> output computes
value-program`, puts six structured `claim-case/v1` observations inside those
value programs, turns that source relation into a fixed claim denominator, and
emits an append-only transition and cause-receipt ledger. Each case declares a
proposition tuple and zero or more independently sourced observation tuples;
the source does not declare a result label or expected status. Both producer
and judge reconstruct the cases from `syntax.ParseFile -> bidir.Lower`; neither
uses a duplicated activity specification or index-assigned terminal state.

## Adopted and rejected principles

### W3C PROV-DM

[PROV-DM](https://www.w3.org/TR/prov-dm/) models provenance around entities,
activities, derivations, and responsibility. We adopt its useful boundary:
the source activity, evidence entity, and resulting receipt are recorded as
provenance-bearing relations. We reject importing the full PROV vocabulary or
its temporal model because this experiment needs a closed, deterministic
three-state lifecycle rather than a general provenance interchange model.

### Belnap, *A Useful Four-Valued Logic*

[Belnap's 1977 paper](https://doi.org/10.1007/978-94-010-1161-7_2) motivates
preserving contradictory information without allowing inconsistency to erase
the knowledge base or cause arbitrary conclusions. We adopt the narrow idea
that contradictory evidence is first-class and must remain inspectable. We
reject copying FDE's four values: this calculus has three lifecycle states
(`OPEN`, `DISCHARGED`, `REFUTED`), while `UNKNOWN` is a resolution of missing,
insufficient, or blocked evidence and `FAIL_CLOSED` is a case decision, not a
claim state.

## Calculus and falsifiers

The transition form is:

```text
proposition + observation tuples + dependency + cause receipt
    -> relation -> (before, after, event, digest)
```

The evaluator compares each observation tuple to the proposition: equality
produces `SUPPORTS`, same subject/predicate with a different object produces
`CONTRADICTS`, no tuple produces `UNAVAILABLE`, and mixed or conflicting tuple
comparisons produce `AMBIGUOUS`; subject/predicate mismatch produces
`INSUFFICIENT`. Only `SUPPORTS` and direct `CONTRADICTS` close a claim. All
other relations preserve `OPEN`; dependency blocking records the upstream
claim without propagating a terminal state. A falsifying input would accept a
missing claim, change a claim without an event, lose a prior event digest,
classify a dependency block as a direct absence, or treat ambiguous evidence
as a closure. The independent judge rejects each of those forms. The top-level
receipt keeps conformance (`PASS`/`FAIL_CLOSED`) separate from subject decision
(`MIXED`/`FAIL_CLOSED`), lower-resolution reason, and subject counts.

The CI falsifiers are fixed at one semantic and one nonsemantic intervention.
Changing the source `observed_object` from `down` to `healthy` changes the
observation comparison, relation, semantic digest, subject outcome, and
transition. Adding only a source comment changes the raw digest while the
lowered semantic digest, transition ledger, and decisions remain equal. A
coherently resealed ledger edit is still rejected because its event no longer
matches source reconstruction. Repository before/after snapshots and the
actual `runtime.Version()` are bound into the receipt rather than reported as
constants.

This experiment does not prove evidence truth, source authority, temporal
freshness, or multi-source conflict arbitration. Those remain open
counterclaims for a later experiment.
