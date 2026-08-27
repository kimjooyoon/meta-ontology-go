# Partial-knowledge composition calculus

This experiment asks a narrow question: when several meta-operations are
composed, can a receipt preserve the difference between an operation that is
directly unknown, an operation blocked by an unresolved dependency, and a
known invariant that does not establish the target claim?

## Research basis

The adopted abstract-interpretation principle comes from P. Cousot and R.
Cousot, “Abstract interpretation: a unified lattice model for static analysis
of programs by construction or approximation of fixpoints,” POPL 1977. Their
model represents program properties in a complete semilattice and requires
order-preserving interpretations, so an approximation may lose precision but
must not manufacture a more precise fact. This experiment adopts that
non-invention rule for composition.

The adopted three-valued principle is the strong Kleene treatment documented
in the Yale/Stanford account of Peirce’s three-valued logic: an unknown input
remains unknown under negation, conjunction, and disjunction when no classical
value is forced. This experiment adopts the analogous rule that direct
`UNKNOWN` remains a non-promotable meta value. It does not treat unknown as
false, true, or a transport error.

The formal sources are:

- [Cousot and Cousot, POPL 1977](https://cs.nyu.edu/~pcousot/COUSOTpapers/POPL77.shtml)
- [Peirce’s three-valued logic, Stanford Encyclopedia of Philosophy](https://plato.stanford.edu/archives/sum2018/entries/peirce-logic/three-valued-logic.html)
- [Truth versus information in logic programming, Cambridge Core](https://www.cambridge.org/core/journals/theory-and-practice-of-logic-programming/article/truth-versus-information-in-logic-programming/FCE4AEFF496594C839719E2EC2A0DEDE)

Rejected principles are equally important. Generic optional/error propagation
is rejected because it collapses knowledge cause into transport mechanics.
“Any non-error is success” is rejected because an invariant can be known while
the target claim remains unproved. A dependency block is not rewritten as a
direct unknown: its receipt preserves the dependency edge and lets a later
consumer locate the missing proof.

## Calculus

The input domain has four atomic values: `EXACT`, `DIRECT_UNKNOWN`,
`DEPENDENCY_BLOCKED`, and `INVARIANT_ONLY`. Composition is a strict product of
knowledge causes:

| left/right evidence | result |
| --- | --- |
| all `EXACT` | `EXACT` |
| any direct unknown, no dependency block | `DIRECT_UNKNOWN` |
| any dependency block, no direct unknown | `DEPENDENCY_BLOCKED` |
| no unknown/block and any invariant | `INVARIANT_ONLY` |
| both direct unknown and dependency block | `MIXED_UNRESOLVED` |

Only `EXACT` is a top-success case. `INVARIANT_ONLY` is retained as useful
knowledge but is `HOLD`; every unresolved state is `FAIL_CLOSED`. The mixed
case carries both cause sets instead of choosing one and hiding the other.

The fixed corpus has five pairwise compositions: one exact result, one direct
unknown, one dependency block, one invariant-preservation result, and one
mixed result. The producer is `partial-knowledge-producer`; the consumer is
`partial-knowledge-composition-consumer`; the operation is named in every
case, and proof choice is one of `FOUNDATION`, `COHERENCE`, or `REGRESSION`.

## Evidence boundary and falsification

The Go producer emits a digest-bound receipt from the checked-in Gooo source
and fixture. A separate verifier re-parses the source and fixture, recomputes
the composition states, recomputes the append-only claim-transition chain,
and checks the producer receipt without importing its evaluator. A changed
case order, expected state, cause, denominator, source, or receipt digest must
fail verification.

The denominator is exactly 5 cases. The expected result is 1/1 exact,
1/1 direct-unknown classification, 1/1 dependency-block classification,
1/1 invariant-preservation classification, 1/1 mixed-cause classification,
and 5/5 non-exact cases not promoted. The repository write count is 0 and
promotion authority is false. This is a read-only proof of the calculus, not a
grant of semantic or repository authority.
