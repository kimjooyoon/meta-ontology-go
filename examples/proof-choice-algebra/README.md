# Proof-choice algebra experiment

This experiment treats the choice of justification as a meta value attached to
each claim and metric. It does not reproduce Lean, Rocq, or another language's
type system. The `.gooo` files remain ordinary source declarations; their
`proof-choice` comments are source-owned evidence metadata consumed by the
experiment.

## Algebra

The three values are the Munchausen-trilemma routes used by this experiment:

| Choice | Meaning in this experiment |
| --- | --- |
| `FOUNDATION` | bind the claim to a stable source or identity anchor |
| `COHERENCE` | show agreement with neighboring claims or relations |
| `REGRESSION` | replay a prior receipt or state and compare it |

For evidence bundles, `⊕` is a disjoint union keyed by stable item ID:

```text
x ⊕ y = x ∪ y       when IDs are disjoint
x ⊕ x = x           for an exact duplicate witness
x ⊕ y = FAIL_CLOSED when one choice is absent/UNKNOWN
x ⊕ y = FAIL_CLOSED when one ID carries conflicting values
```

Every claim and metric must carry exactly one of the three choices. Producer,
consumer, and meta-operation are required routing fields. Every metric uses the
fixed denominator `3`; a different denominator is not normalized or averaged.
A persistent claim transition must retain the claim ID and choice while moving
from `ASSERTED` to `PERSISTED`. `UNKNOWN` stage, step, or reason is evidence
failure, not a fourth proof route.

The producer emits a deterministic receipt with source and receipt digests,
per-item indicators, fixed-denominator summary counts, producer/consumer/meta
operation fields, persistent transitions, and:

```json
{"repository_writes":0,"mutation_authority":false}
```

The independent judge decodes this wire shape itself and recomputes the verdict
without importing the producer evaluator. It checks the receipt digest before
accepting the producer's decision.

## Cases

- `foundation.gooo`, `coherence.gooo`, and `regression.gooo` are one-choice
  positive cases.
- `combined.gooo` demonstrates disjoint composition of all three routes.
- `missing-choice.gooo` omits a claim choice and also preserves `UNKNOWN`
  context; it must fail closed.
- `contradiction.gooo` assigns two different choices to one claim ID; it must
  fail closed.
- `unknown-context.gooo` makes stage, step, and reason unknown; it must fail
  closed even though its choices are otherwise valid.

The CI workflow produces and replays receipts, then invokes the independent
judge on every case. It is the validation path for this experiment; no source
or repository mutation is authorized by the tool.

## Adopted and rejected principles

The design adopts two official proof-tool principles:

1. The [Rocq Reference Manual](https://rocq-prover.org/doc/master/refman/index.html)
   says a small kernel performs final verification of tactic-generated proof
   output. We adopt the separation between a producer and a smaller, explicit
   checker; we reject treating a successful producer run as self-authenticating.
2. The [Lean Language Reference](https://lean-lang.org/doc/reference/latest/)
   describes a minimal kernel that checks proof terms and notes that tactics
   produce terms checked by that kernel. We adopt independently checkable,
   digest-bound receipts; we reject copying Lean's proof terms, elaboration, or
   dependent types because this experiment is about selecting evidence routes,
   not expressing object-level propositions.

These sources justify the checker boundary, not the truth of the three
Munchausen labels. The labels and combination laws are an explicit, falsifiable
experiment contract owned by the `.gooo` fixtures and this package.

## Falsification probes

The hypothesis is falsified operationally if any of these mutations still pass:

1. remove or set a choice to `UNKNOWN`;
2. duplicate an ID with another choice;
3. change a metric denominator from `3`;
4. set stage, step, or reason to `UNKNOWN`;
5. alter a receipt without recomputing its digest; or
6. alter a persistent transition's claim ID or choice.

The experiment measures proof-route integrity and provenance routing. It does
not establish that a claim is true, nor does it claim that one route is
epistemically superior to the others.
