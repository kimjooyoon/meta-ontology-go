# Deterministic ambiguity budget

This experiment measures interpretation ambiguity as an integer set, not as
probability or natural-language confidence. It is a read-only meta-operation:
the repository effect is fixed at zero writes and zero mutation authority.

## Executable source and rule

The real [`.gooo` source](../../examples/ambiguity-budget/main.gooo) is the
observation authority. Its canonical `computes` declarations contain the
fixed budget `(2,1,2)` and all four observed case sets:

| case | observed set | derived class / subject result / claim |
| --- | --- | --- |
| zero | `(1,0,1)` | `ZERO / PASS / EXACT / OPEN→DISCHARGED` |
| boundary | `(2,1,2)` | `BOUNDARY / PASS / EXACT / OPEN→DISCHARGED` |
| over | `(3,2,3)` | `OVER / FAIL_CLOSED / LOWER_RESOLUTION / OPEN→REFUTED` |
| unknown | `(2,?,2)` | `UNKNOWN / UNKNOWN / LOWER_RESOLUTION / OPEN→OPEN` |

The coordinates count interpretation candidates, unresolved branches, and
evidence paths. A known set is within the budget only when every coordinate is
within its corresponding limit. Any excess descends to lower resolution. Every
claim ledger row starts at `OPEN`; a sufficient observation discharges it, a
real excess refutes it, and an observation defect preserves `OPEN`. `EXACT`
and `LOWER_RESOLUTION` are resolution values only. The `?` coordinate is not a
value or a label: it is an unobserved `unresolved_branches` coordinate
discovered by the evaluator.

The JSON [contract](../../examples/ambiguity-budget/contract.json) contains
only stable source identity, case IDs, intervention IDs, and the fixed
denominator. It intentionally does not repeat the budget, case counts, class,
decision, or reason as evidence. The producer parses with
`syntax.ParseFile -> bidir.Lower`, validates the canonical semantic IR, and
then reads the lowered `computes` programs. The independent verifier repeats
that parse/lowering and rule evaluation without importing the producer.

`CONFORMANCE_DECISION` describes whether the receipt matches the executable
source and contract. `SUBJECT_DECISION` describes the case vector. Therefore a
receipt can conform while the subject remains `MIXED / LOWER_RESOLUTION`; an
expected `UNKNOWN` is not converted into an aggregate `PASS / EXACT`.

Every persistent claim transition records `stage`, `step`, `reason`, and an
evidence digest. The receipt carries the producer, consumer, meta-operation,
proof choice, source/semantic digests, and zero-effect guard. Interventions
carry their before/after claim transitions explicitly. There is no aggregate
score: the fixed denominator is `2`, counting the two declared interventions,
not the twelve integer coordinates.

## Adopted and rejected principles

Scott and Johnstone’s paper, [“Recognition is not parsing — SPPF-style parsing
from cubic recognisers”](https://doi.org/10.1016/j.scico.2009.07.001), presents
SPPF-style representations for all possible derivations while sharing common
structure. The experiment adopts the accounting boundary that alternatives
remain observable objects until a later operation resolves them. It rejects
an SPPF implementation here because the `.gooo` parser is a deterministic
stage-0 parser and this experiment needs an independent integer ledger.

Economopoulos’s [*Generalised LR parsing algorithms* thesis](https://ir.cwi.nl/pub/14233/14233B.pdf), especially §§4.4–4.4.1,
describes reduce/reduce conflicts, local ambiguity, and packing multiple
derivations into a shared forest. The experiment adopts explicit branch and
evidence-path accounting. It rejects treating a shift/reduce choice as
probability, confidence, or evidence of semantic correctness, and it does not
make parser-specific graph structures an implementation dependency.

The adopted principle is: make alternatives and the cost of resolving them
explicit as integer observations and meta-operations. Rejected principles are
probability, confidence prose, default disambiguation, intent recognition,
and semantic-correctness claims. The `not_claimed` array makes the boundary
machine-readable.

## Interventions and falsification

The semantic intervention increments the boundary case’s candidate count from
`2` to `3`; the source and semantic digests change and the subject descends to
`FAIL_CLOSED / LOWER_RESOLUTION`. The nonsemantic intervention inserts only a
comment; the source digest changes while the lowered semantic digest, counts,
and subject result remain unchanged. Both are replayed in memory and recorded
in the receipt.

The claim is falsified by nondeterministic replay; an over-budget case that
remains exact; an unknown case that loses its coordinate or `OPEN` transition;
duplicated observed counts in the contract; a forged source, semantic, or
receipt digest accepted by the judge; a producer import in the judge; a changed
denominator; or any non-zero repository effect.
