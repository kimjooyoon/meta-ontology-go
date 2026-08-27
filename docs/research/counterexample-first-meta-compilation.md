# Counterexample-first meta compilation

Status: bounded philosophy experiment, not a general compiler feature.

## Claim

The experiment tests a causal ordering claim:

```text
success example is context
        ↓
minimum failing counterexample is required input
        ↓
resolution evidence must bind to that counterexample
        ↓
only then may the meta compiler raise COMPILE_DECISION
```

The relation is observable because the decision receipt contains the
counterexample ID and digest, the resolution ID and digest, and three retained
claim transitions. The success example is included in `decision_input` only as
a non-authoritative digest. The independent judge recomputes the expected
receipt from the source, scenario corpus, and contract; it does not accept a
producer decision merely because its receipt is self-consistent.

## What the references contribute

- [QuickCheck `shrink` API documentation](https://hackage.haskell.org/package/QuickCheck/docs/Test-QuickCheck.html)
  defines immediate shrink candidates and recommends type-specific shrinking.
  It also warns that recursive or invariant-bearing data needs a shrinker that
  preserves its invariants. This experiment therefore records a shrink trace
  and requires every step to strictly decrease the pinned size while retaining
  the failure.
- [QuickChick's Software Foundations chapter](https://deepspec.github.io/sf/qc-current/QC.html)
  describes greedy shrinking from a known failing value until every immediate
  shrink succeeds, yielding a *locally* minimal counterexample. It states the
  key limitation: the shrinker must be strictly decreasing in a total order.
  Our `minimal` claim is deliberately local; it is not global minimality.
- [Isabelle's official Nitpick manual](https://www.isabelle.in.tum.de/website-Isabelle2021/dist/Isabelle2021/doc/nitpick.pdf)
  presents counterexample generation as bounded model finding and documents
  scope, timeout, and potential-counterexample limitations. We consequently
  preserve `UNKNOWN` stage/step/reason and lower resolution instead of treating
  absence of a model or an unverified coordinate as proof of safety.

## Principle and boundary

The useful principle is not "a counterexample exists". It is that the
counterexample changes the meta decision's admissible inputs: no counterexample
means `COUNTEREXAMPLE_REQUIRED`; a non-minimal one means
`COUNTEREXAMPLE_NOT_MINIMAL`; a minimal but unresolved one means
`COUNTEREXAMPLE_UNRESOLVED`; only a minimal counterexample with accepted,
bound resolution evidence yields `COUNTEREXAMPLE_RESOLVED`.

This does not prove the compiler is correct. A bounded corpus can miss a
different failing shape, a custom shrink order can miss a smaller witness, and
resolution evidence can only establish the contract it names. Those are the
experiment's explicit falsification paths.

## Fixed denominator and evidence

The contract fixes five cases, ten indicators, and fifteen claim transitions
under `counterexample-first-denominator/v1`. It records producer,
consumer, meta-operation, proof choice, source digest, effects, and an
independent-judge digest. The semantic operation is read-only:
`repository_writes=0` and `mutation_authority=false`. Replaying the producer
and judge on the same exact head must produce byte-identical receipts and
reports.
