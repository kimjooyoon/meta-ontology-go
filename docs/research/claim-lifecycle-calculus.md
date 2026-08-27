# Claim lifecycle calculus research note

## Question

Can a Gooo meta-operation preserve an assertion while evidence changes its
state, including an explicit refutation, without treating missing evidence as
false or silently dropping the assertion?

The experiment is deliberately smaller than a language feature. It observes
the existing Gooo source relation `activity(inputs) -> output computes
value-program`, turns that relation into a fixed claim denominator, and emits
an append-only transition and cause-receipt ledger.

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
(`OPEN`, `DISCHARGED`, `REFUTED`), while `UNKNOWN` is a resolution of missing
or blocked evidence and `FAIL_CLOSED` is a case decision, not a claim state.

## Calculus and falsifiers

The transition form is:

```text
claim + evidence + cause receipt -> (before, after, event, digest)
```

The intended closure is `OPEN -> DISCHARGED` for supporting evidence and
`OPEN -> REFUTED` for contradicting evidence. A falsifying input would accept
a missing claim, change a claim without an event, lose a prior event digest,
classify a dependency block as a direct absence, or treat ambiguous evidence
as a closure. The independent judge rejects each of those forms.

This experiment does not prove evidence truth, source authority, temporal
freshness, or multi-source conflict arbitration. Those remain open
counterclaims for a later experiment.
