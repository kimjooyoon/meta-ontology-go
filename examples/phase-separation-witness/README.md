# Phase separation witness

This experiment makes a meta-level boundary executable as Gooo values. The
three phase-local value spaces are `source`, `expansion`, and `execution`.
Values, authority, and evidence do not cross those spaces implicitly. The only
cross-phase protocol is an explicit `claim` transfer along the adjacent edges
`source -> expansion -> execution`; the claim transition records
`DECLARED -> PRESERVED` without moving the phase-local value itself.

`main.gooo` is the one clean case. `leaks.gooo` contains five fixed
counterexamples: value, authority, and evidence crossings, a source-to-execution
skip, and a reverse edge. `unknown.gooo` is intentionally malformed so the
receipt preserves an `UNKNOWN` stage/step/reason instead of guessing.

The producer emits a deterministic receipt. The separate
`phase-separation-adjudicator` command decodes that receipt using its own
receipt types and independently checks its digest, fixed denominators, case
reasons, preserved claim transitions, and zero authority. It is deliberately
not a second call to the producer's evaluator.

Fixed coordinates are 1/1 clean, 5/5 leakage catches, 2/2 preserved claim
transitions, and 12/12 indicators. The producer and consumer identities,
meta-operation, and proof choice are recorded in every view. The default
toolchain is Go 1.27.0; repository writes and mutation/promotion authority are
zero/false.

## Adopted and rejected principles

The experiment adopts the boundary idea from the official Racket Guide's
[General Phase Levels](https://docs.racket-lang.org/guide/phases.html): phases
are separate computations and communicate through an explicit code-producing
protocol. It also adopts the Racket Reference's [Evaluation
Model](https://docs.racket-lang.org/reference/eval-model.html), which defines
phase 0 as execution and explains that expansion-time bindings are separate
from execution-time bindings. The [Syntax
Model](https://docs.racket-lang.org/reference/syntax-model.html) supplies the
additional principle that bindings are phase-specific.

This experiment rejects two tempting but unsafe reductions: treating a value
that happens to have the same spelling as the same value at another phase, and
treating evidence or authority as ordinary data that can be copied through a
phase edge. It also does not claim to implement Racket's macro expander,
module instantiation, or a general Gooo expression language; it isolates the
falsifiable phase-boundary witness as a data-only `.gooo` value experiment.
