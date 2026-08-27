# Deterministic ambiguity budget

This PR records a bounded experiment for interpreting ambiguity in semantic
meta-operations. It does not add a confidence score and does not infer intent
from natural language.

## Decision rule

The observation is an integer set:

`A = (c, b, e)`

where `c` is the number of interpretation candidates, `b` is the number of
unresolved branches, and `e` is the number of evidence paths that must be
reconciled. The contract supplies a fixed upper bound `B = (2, 1, 2)`.

For a known observation, the deterministic rule is:

`A <= B` coordinate-by-coordinate → `PASS / EXACT`;

`A_i > B_i` for any coordinate → `FAIL_CLOSED / LOWER_RESOLUTION`.

An unknown observation is `UNKNOWN / LOWER_RESOLUTION`, with an explicit
`stage`, `step`, and `reason`. It does not erase the claim transition that led
to the lower-resolution result. The receipt has a fixed denominator of `12`:
four cases × three integer coordinates. This is a count of obligations, not a
percentage or confidence estimate.

## Adopted and rejected principles

Scott and Johnstone’s paper, [“Recognition is not parsing — SPPF-style parsing
from cubic recognisers”](https://doi.org/10.1016/j.scico.2009.07.001), presents
SPPF-style representations for all possible derivations while sharing common
structure. The experiment adopts the key accounting boundary: alternatives
remain observable objects until a later operation resolves them. It does not
adopt an SPPF implementation because the current `.gooo` syntax parser is a
deterministic stage-0 parser and this PR is explicitly an independent
read-only meta-operation.

Economopoulos’s [*Generalised LR parsing algorithms* thesis](https://ir.cwi.nl/pub/14233/14233B.pdf), especially §§4.4–4.4.1,
describes reduce/reduce conflicts, local ambiguity, and packing multiple
derivations into a shared forest. The experiment adopts explicit branch and
evidence-path accounting and rejects an implicit shift/reduce choice as
evidence that one interpretation is more probable. It also rejects the
thesis’s parser-specific graph structures as an implementation dependency:
the receipt records only the deterministic integer boundary needed by this
experiment.

Thus the adopted principle is “make alternatives and the cost of resolving
them explicit.” Rejected principles are probability, confidence prose,
default disambiguation, and semantic-correctness claims. The `not_claimed`
array makes those boundaries machine-readable.

## Evidence contract

The producer reads the real [ambiguity-budget `.gooo` source](../../examples/ambiguity-budget/main.gooo), binds its source digest and declaration counts, and
emits a data-only receipt. The independent verifier does not import the
producer package: it re-reads the source, recomputes its digest and declaration
counts, recomputes the decision rule, checks all twelve coordinate indicators,
checks the claim-transition list, and validates the receipt digest.

The producer, consumer, meta-operation, and proof choices are explicit on the
receipt and every indicator. Repository writes and mutation authority are
fixed at zero/false. The four cases are zero (`1/0/1`), boundary (`2/1/2`),
over (`3/2/3`), and unknown. The over and unknown cases are expected
counterexamples: the whole experiment can pass only when those lower-resolution
transitions are observed exactly.

## Falsification

The claim is falsified by any nondeterministic replay; by an over-budget case
remaining exact; by an unknown case losing its coordinate or transition; by a
changed denominator; by a forged source digest, receipt digest, or transition
accepted by the independent verifier; or by any non-zero repository effect.

