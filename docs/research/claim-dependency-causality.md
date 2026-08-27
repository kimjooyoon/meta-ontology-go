# Claim dependency causality

## Experiment question

When a meta claim is unresolved, can a receipt distinguish the local failure
that made the claim `UNKNOWN` from a downstream claim that is only blocked by
that open predecessor? The experiment answers this with a fixed, six-claim
Gooo contract. It does not treat every graph edge as a failure: the producer
records the local observation, propagates the state, and assigns responsibility
to the root or to the upstream claim.

## External principles used

The design takes three narrow ideas from published, authoritative material:

1. W3C PROV-O models provenance with entities, activities, and agents. Its
   `used`, `wasGeneratedBy`, `wasInformedBy`, and `wasDerivedFrom` relations
   construct chains, while qualified relations can add detail to an influence.
   This experiment therefore names a producer, consumer, operation, coordinate,
   and root transition digest rather than treating a bare edge as evidence.
   See [PROV-O](https://www.w3.org/TR/prov-o/).
2. OpenLineage makes the metadata producer explicit, defines event transitions
   for a run, and distinguishes design lineage from run-observed lineage. This
   experiment follows that separation: the fixed graph is declared contract
   metadata, while the twelve transition events are the observed run record.
   See [OpenLineage object model](https://openlineage.io/docs/spec/object-model/),
   [run cycle](https://openlineage.io/docs/spec/run-cycle/), and
   [producer field](https://openlineage.io/docs/spec/producers/).
3. Causal graphical models use directed paths to represent ancestry and direct
   causes relative to a chosen variable set. This experiment borrows only the
   directional path discipline: `CausePath` is the shortest path in the fixed
   claim contract, with a declared shortcut for the decision claim. It does not
   borrow statistical, intervention, or counterfactual semantics. See the
   [Stanford Encyclopedia of Philosophy entry on causal models](https://plato.stanford.edu/entries/causal-models/).

## Contract and state rule

The fixed nodes are `source-observed`, `producer-bound`, `proof-choice-bound`,
`consumer-bound`, `read-only-bound`, and `decision-replay-bound`. The fixed edge
contract is eight edges, including `C1 -> C2 -> C6` as the shortest route to the
decision claim and three longer proof/consumer/authority routes into that claim.

The rule is deliberately local:

- A claim with unavailable local evidence is `DIRECT_UNKNOWN`, remains `OPEN`,
  and owns the failure.
- A claim whose local evidence is not the root cause but has an open incoming
  predecessor is `DEPENDENCY_BLOCKED`, remains `OPEN`, and points to the root
  evidence plus its immediate blocking frontier.
- A contradictory local observation is `DIRECT_REFUTED`; the same state on a
  dependent claim is `DEPENDENCY_REFUTED`.
- When the root observation is repaired, it becomes `DISCHARGED`; dependent
  claims become `DEPENDENCY_RECOVERED` along their minimum paths.

The transition ledger is fixed at twelve events per receipt: six registrations
(`UNRECORDED -> OPEN`) and six outcome transitions. The three receipts exercise
the final states `OPEN`, `REFUTED`, and `DISCHARGED`. The receipt digest and
transition chain make omission, reordering, or resealing observable.

## Falsifiable predictions and limitations

The experiment is falsified if an independent judge accepts a changed producer,
consumer, proof choice, coordinate, graph edge, transition, cause path, or
decision; if the unknown receipt does not report exactly `1` direct unknown,
`5` blocked claims, `8` blocking edges, and maximum depth `2`; or if the
recovered receipt does not report `6/6` discharged claims and `5` unique
recovery edges. CI runs the producer twice and the independent judge twice and
requires byte-identical receipts.

The result is intentionally weaker than provenance or causal inference. It
does not establish that the graph is complete for arbitrary Gooo programs, that
an arrow is a real-world cause, that evidence is true, or that a downstream
claim could never have independent evidence. It assumes an acyclic, closed,
hand-declared six-claim contract and uses source markers to select cases. A
future experiment must add explicit competing roots, cycles, independent
downstream evidence, and a changed contract before generalizing the rule.

The only authority demonstrated is a read-only meta decision: repository writes
and semantic promotion remain zero.
