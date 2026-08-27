# Self-improvement termination witness

This experiment observes repeated executions of the self-improvement
metaprogram as a finite state trace. It proves only what the trace supports;
it does not promote every completed observation to `FIXED_POINT` and it never
writes the repository or authorizes promotion.

The input contract fixes the denominator at ten indicators and binds the
producer, consumer, meta-operation, proof choice, stage, step, and reason.
Each state is a content digest. A `NO_CHANGE` observation is a self-transition
whose before and after digests are equal. A non-adjacent repeated digest is a
cycle after stuttering no-change observations are removed.

`state count` is the length of the state sequence after adjacent no-change
stuttering is removed; it includes the repeated endpoint that proves a cycle.

| case | observed trace | state count | repeated states | decision | termination claim |
| --- | ---: | ---: | ---: | --- | --- |
| `fixed-point.json` | 1/4 steps | 1 | 0 | `FIXED_POINT` | proven from the final no-change self-transition |
| `cycle-2.json` | 2/4 steps | 3 | 1 (period 2) | `CYCLE` | refuted by a period-2 repeated state |
| `in-progress.json` | 2/4 steps | 3 | 0 | `IN_PROGRESS` | no terminal or cycle evidence yet |
| `divergence-possible.json` | 2/2 steps | 3 | 0 | `DIVERGENCE_POSSIBLE` | only a bounded strictly-growing prefix; not a proof of infinity |

Every valid receipt has `10/10 = 10000` basis points because the denominator
measures whether the selected branch was observed and bound, not whether the
branch was a fixed point. Only the fixed-point branch sets
`summary.termination_proven` to `true`.

## Formal design choices

The experiment uses these sources:

1. [Baader and Nipkow, *Term Rewriting and All That*, Chapter 5:
   Termination](https://www.cambridge.org/core/books/abs/term-rewriting-and-all-that/termination/36ACECF2B0D53A6933FA6EDAF7E219FF)
   treats termination as a property of rewrite systems, notes its general
   undecidability, and introduces reduction orders as a way to prove the
   property.
2. [Lean 4 Reference, Recursive Definitions](https://lean-lang.org/doc/reference/latest/Definitions/Recursive-Definitions/)
   defines well-founded recursion through a measure that decreases at every
   recursive call and explains that a well-founded relation has no infinite
   descending chain.
3. [Cornell CS 4110, Fixed Points](https://www.cs.cornell.edu/courses/cs4110/2014fa/lectures/slides07.pdf)
   defines a fixed point by `F(x) = x` and describes iterative approximation;
   that equation alone is not an operational termination proof.

Adopted principles:

- Accept `FIXED_POINT` only for the final `NO_CHANGE` self-transition after
  cycle analysis. Equality is the local fixed-point witness.
- Preserve a state chain and a fixed step budget. A strictly increasing rank at
  the budget boundary yields `DIVERGENCE_POSSIBLE`, never `FIXED_POINT`.
- Treat a non-adjacent repeated state as `CYCLE`, including a concrete 2-cycle.
- Keep `IN_PROGRESS` when the finite prefix has neither terminal nor cycle
  evidence. The receipt records an exact observation, not an infinite claim.
- Emit a claim transition from `UNPROVEN` to `OBSERVED` and then to the exact
  observed decision. The independent judge recomputes this transition.

Rejected shortcuts:

- A run limit is not a termination proof; it is only a boundary for a
  `DIVERGENCE_POSSIBLE` observation.
- A prior `FIXED_POINT` label is not trusted as evidence. The before/after
  digests and `NO_CHANGE` reason must agree.
- A repeated state is not automatically a fixed point: period two is a cycle,
  and only adjacent no-change stuttering is ignored by the cycle scan.
- A fixed-point theorem about a recursive functional does not establish that a
  concrete self-improvement workflow reaches that point.

The producer is `selfimprovementtermination.Evaluate`. The consumer is
`self-improvement-cycle`, and the meta-operation is
`prove-self-improvement-termination` under the `TERMINATION` proof choice. The
independent command uses a separate verifier implementation and checks the
receipt digest, fixed denominator, trace, branch, and claim transitions.
