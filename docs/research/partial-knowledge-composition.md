# Partial-knowledge composition calculus

This experiment asks whether a receipt can compose several meta-operation
observations without confusing unknown knowledge, dependency blockage, and a
known invariant. The source of truth is `main.gooo`: five upstream observation
receipts are encoded in `activity ... computes` programs. The JSON output is a
derived artifact, not a case fixture.

## Source and reconstruction boundary

Each observation contains `required`, `observed`, `observed_available`,
`dependency_claim_id`, `prior_state`, and `invariant_evidence`. Neither the
source nor the input model contains a final `state`, `decision`, or
`resolution`. Both the producer and the independent verifier call
`syntax.ParseFile` followed by `bidir.Lower`, check the lowered activity
`ValueProgram`, and derive the five cases from the lowered source. The
verifier has its own parser, lowering, composition, digest, and receipt
reconstruction path and does not import the producer package.

## Research basis

The adopted abstract-interpretation principle comes from P. Cousot and R.
Cousot, “Abstract interpretation: a unified lattice model for static analysis
of programs by construction or approximation of fixpoints,” POPL 1977. Its
complete-semilattice and order-preserving viewpoint supports an approximation
that loses precision without manufacturing a more precise fact. This
experiment adopts that non-invention rule for composition.

The adopted three-valued principle is the strong Kleene treatment documented
in the Stanford Encyclopedia account of Peirce’s three-valued logic: an
unknown value is not silently converted into true or false by the connective.
This experiment analogously keeps a directly unavailable observation as an
open claim at lower resolution.

The formal sources are:

- [Cousot and Cousot, POPL 1977](https://cs.nyu.edu/~pcousot/COUSOTpapers/POPL77.shtml)
- [Peirce’s three-valued logic, Stanford Encyclopedia of Philosophy](https://plato.stanford.edu/archives/sum2018/entries/peirce-logic/three-valued-logic.html)
- [Truth versus information in logic programming, Cambridge Core](https://www.cambridge.org/core/journals/theory-and-practice-of-logic-programming/article/truth-versus-information-in-logic-programming/FCE4AEFF496594C839719E2EC2A0DEDE)

Rejected principles are also explicit. Generic optional/error propagation is
rejected because it describes transport rather than knowledge resolution.
“Any non-error is success” is rejected because an invariant can be known while
the target claim remains unproved. A dependency block is not rewritten as a
direct unknown: its dependency edge remains available to a later consumer.

## Calculus and claim semantics

The atomic evidence rules are:

| Evidence | Derived state | Outcome | Claim transition |
|---|---|---|---|
| observations available and equal to required | `EXACT` | `PASS` / `EXACT` | `OPEN -> DISCHARGED` |
| observation unavailable without dependency | `DIRECT_UNKNOWN` | `UNKNOWN` / `LOWER_RESOLUTION` | `OPEN -> OPEN` |
| observation unavailable with dependency claim | `DEPENDENCY_BLOCKED` | `UNKNOWN` / `LOWER_RESOLUTION` | `OPEN -> OPEN` |
| available observation with invariant evidence | `INVARIANT_ONLY` | `HOLD` / `INVARIANT_ONLY` | `OPEN -> OPEN` |
| both direct unknown and dependency block | `MIXED_UNRESOLVED` | `UNKNOWN` / `LOWER_RESOLUTION` | `OPEN -> OPEN` |

Only `EXACT` is a case-level top-success value. The receipt-level decision is
`CALCULUS_PROVEN`, with `resolution=CALCULUS` and
`subject_resolution=PARTIAL_KNOWLEDGE`; these fields make the proof of the
rule replay separate from the subject-case result counts. The fixed corpus is
5/5 source-derived cases: one of each row above. It yields 1 exact case, 4
non-exact cases, 4/4 non-exact cases not promoted, 1 discharged claim, 4 open
claims, and 5 digest-linked transitions.

Every case and claim records producer `partial-knowledge-producer`, consumer
`partial-knowledge-composition-consumer`, a meta-operation, proof choice,
stage, step, reason, source activity, semantic IR digest, observation digest,
and evidence digest.

## Interventions and falsifiability

The semantic intervention changes `direct-unknown.left.observed_available`
from `false` to `true` and supplies the required observation. It must change
that case from `UNKNOWN` / `LOWER_RESOLUTION` with `OPEN -> OPEN` to
`PASS` / `EXACT` with `OPEN -> DISCHARGED`; this is semantic causality 1/1.
A comment-only source intervention changes the raw source digest but must keep
the lowered semantic IR digest and semantic projection identical; this is
nonsemantic preservation 1/1.

The Action wrapper exposes the following fixed metrics: producer imports
0/0, source cases 5/5, semantic causality 1/1, nonsemantic preservation 1/1,
and open claims preserved 4/4. It uploads both JSON and a human-readable
Markdown artifact. Any changed computes field, source activity, lowering
result, case order, dependency edge, claim transition, digest, denominator,
or authority field must make the verifier fail. Repository writes remain 0 and
promotion authority remains false.
