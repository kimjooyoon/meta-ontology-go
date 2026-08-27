# Deterministic ambiguity budget

This is a read-only philosophy experiment. It treats ambiguity as an integer
set, not as natural-language confidence:

`(interpretation_candidates, unresolved_branches, evidence_paths)`

The fixed budget is `(2, 1, 2)`. A known set at or below every bound remains
`PASS / EXACT`; a known set over any bound becomes `FAIL_CLOSED /
LOWER_RESOLUTION`. An `UNKNOWN` input also lowers resolution and retains its
`stage`, `step`, `reason`, and claim transition.

The source is [main.gooo](main.gooo). The producer parses and binds that source
to [contract.json](contract.json), then emits a deterministic receipt. The
independent verifier re-derives the three integer coordinates, source binding,
case transitions, fixed denominator (`12`), and zero-write effect.

## Evidence choices

The design adopts the “retain alternatives as a finite structural object” idea
from two formal parsing sources:

- Scott and Johnstone, “Recognition is not parsing — SPPF-style parsing from
  cubic recognisers,” *Science of Computer Programming* (2010),
  [publisher record and DOI](https://doi.org/10.1016/j.scico.2009.07.001).
  It motivates an SPPF containing all possible derivations while sharing common
  substructure. We adopt the separation between alternatives and a chosen
  interpretation; this experiment counts alternatives instead of assigning
  them weights.
- Economopoulos, *Generalised LR parsing algorithms* (Royal Holloway thesis,
  2006), [institutional PDF](https://ir.cwi.nl/pub/14233/14233B.pdf), §§4.4–4.4.1.
  It describes reduce/reduce conflicts, local ambiguity, and packed forests as
  a compact representation of multiple derivations. We adopt explicit
  conflict/branch accounting and reject silently selecting one derivation as a
  confidence claim.

We reject probability, “likely”/“unlikely” labels, default disambiguation, and
implicit tie-breaking: none is present in the JSON schema or decision
procedure. The budget is a policy cap, not a semantic truth score. The
experiment also does not claim that the current deterministic `.gooo` parser
constructs a general SPPF; it only imports the formal accounting principle
into a read-only meta-operation.

## Cases and falsification

The four fixed cases are zero ambiguity (`1/0/1`), boundary (`2/1/2`), over
budget (`3/2/3`), and unknown input. The claim is falsified if a permutation of
the same input changes a receipt, if any over-budget coordinate stays `EXACT`,
if an `UNKNOWN` case loses its coordinate or transition, if the denominator is
not `12`, or if the verifier accepts a modified digest/receipt.
