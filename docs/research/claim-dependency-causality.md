# Claim dependency causality

## Question and boundary

This is a read-only Gooo meta experiment about failure responsibility. It asks
whether a direct `UNKNOWN` observation can be separated from a dependent claim
that is blocked by an unresolved predecessor, while preserving the minimum
causal path. It is not a general dependency engine and it makes no claim about
truth in the external world.

## Source-grounded construction

Both producer and independent consumer begin with raw `.gooo` source and run
`syntax.ParseFile` followed by `bidir.Lower`. They validate the resulting
canonical semantic IR and derive the six claim nodes from its six activity
declarations. The graph edges are not a receipt-only contract: each edge comes
from an IR `wasGeneratedBy` output entity joined to a downstream IR `used`
input entity. The downstream activity's semantic `computes` value is parsed as
one of four closed edge predicates: `SUPPORTS`, `REQUIRES`, `CONTRADICTS`, or
`FAILURE_ENTAILMENT`.

The root activity's semantic value is either `claim.observe:recoverable` or
`claim.observe:contradiction`. A separately digested observation must agree
with that source predicate: `UNKNOWN` has no evidence, `EVIDENCE_ACCEPTED`
requires evidence, and `EXPLICIT_CONTRADICTION` requires both the contradiction
source predicate and evidence. Thus an integer operation or source substring
does not itself mean refutation.

The fixed denominator is six claims, eight typed edges, and twelve initial
transitions. The example graph has two `SUPPORTS`, three `REQUIRES`, two
`CONTRADICTS`, and one `FAILURE_ENTAILMENT` edge. Root-to-derived is a real
semantic relation through the generated/used `RootState` entity.

## State algebra and responsibility

The propagation rule is intentionally asymmetric:

* A direct `UNKNOWN` observation is `OPEN` and `DIRECT_UNKNOWN`.
* An upstream `UNKNOWN` makes a dependent `OPEN` and `DEPENDENCY_BLOCKED`.
* An upstream `REFUTED` on `SUPPORTS` or `REQUIRES` also leaves the dependent
  `OPEN` and blocked; those relations do not entail falsity.
* A dependent becomes `REFUTED` only when an upstream `REFUTED` state reaches it
  through a `CONTRADICTS` or `FAILURE_ENTAILMENT` edge.
* `OPEN -> DISCHARGED` is legal only for a matching `EVIDENCE_ACCEPTED`
  predicate. Every transition retains stage, step, reason, evidence digest,
  provenance, and its predecessor digest.

For `refuted.gooo`, the exact result is one direct refutation, three open
claims, two dependency refutations, five blocking edges, and two effective
refuting edges. The `CONTRADICTS` edge from Root to `ContradictionCheck` and the
`FAILURE_ENTAILMENT` edge from `ContradictionCheck` to
`FailureEntailmentCheck` are the only refuting path in that fixture.

## Append-only recovery

Recovery consumes the prior UNKNOWN receipt, not merely a label. It verifies
the prior receipt digest, its transition head, its twelve-transition chain,
its six `OPEN` claim states, and the current graph digest. The recovered receipt
copies the prior transition prefix byte-for-byte and appends six transitions
from the prior states to `DISCHARGED`; it records the prior receipt digest,
previous transition digest, prior claim states, and the new observation digest.
The resulting exact chain is twelve preserved transitions plus six appended
transitions, with eight recovery edges.

## External principles and limits

The provenance boundary follows [W3C PROV-O](https://www.w3.org/TR/prov-o/):
entities, activities, agents, and typed influence relations provide a useful
vocabulary for source, producer, consumer, and lineage. [OpenLineage's object
model](https://openlineage.io/docs/spec/object-model/) and [run
cycle](https://openlineage.io/docs/spec/run-cycle/) motivate separating a
declared graph from observed transition events; its [producer
field](https://openlineage.io/docs/spec/producers/) motivates explicit
producer identity. The directional minimum-path discipline is informed by
[Stanford's causal-models overview](https://plato.stanford.edu/entries/causal-models/),
but this artifact does not claim statistical causation or counterfactual
identification.

The design is falsified if the consumer accepts a changed graph, edge kind,
observation digest, predecessor digest, transition, stage/step/reason,
provenance, cause path, or decision; if ordinary support/requirement edges
turn an upstream refutation into a dependent refutation; if an unknown state
is silently discharged; or if the exact CI counts change. Semantic value and
edge-type interventions must change the relevant digest and state/propagation
outcome. A comment-only intervention must preserve semantic digest, graph
digest, state vector, and decision.

The experiment assumes an acyclic, closed six-activity fixture and a small
observation vocabulary. It does not establish graph completeness, external
truth, runtime correctness, independent evidence for every downstream claim,
or the adequacy of the four edge types outside this fixture. Read-only effects
are part of the observation contract; repository writes and semantic promotion
authority remain zero.
