# Invariant-preserving transformation experiment

This source contract treats transformation authority as a four-part invariant:

`precondition ∧ transformation ∧ postcondition ∧ regression witness`

The producer emits a data-only receipt. The independent judge recomputes the
decision from the fixed denominator and the receipt, so a producer cannot mint
authority by changing the final decision field. `OPEN` means evidence is absent,
`DISCHARGED` means the obligation has a bound witness, and `REFUTED` records a
counterexample. Only four fixed cases are admitted: a preserved translation, a
semantic violation, missing regression evidence, and an approved artifact.

The approved artifact case records one separate `APPROVED_ARTIFACT_RECORDED`
effect. It still has `repository_writes=0` and `mutation_authority=false`:
recording an approved product is not permission to mutate the repository.

## Research basis and limits

This experiment is informed by [Necula, “Translation Validation for an
Optimizing Compiler,” PLDI 2000](https://dl.acm.org/doi/10.1145/349299.349314),
which validates the actual output of each compilation and uses simulation
relations as witnesses, and [Sultana and Thompson, “Mechanical Verification of
Refactorings”](https://kar.kent.ac.uk/23959/), which formalizes refactorings in
Isabelle/HOL and treats behavior preservation as the correctness condition.

Those works do not make this small Gooo model a verified refactoring engine.
Program equivalence is undecidable; a finite receipt can only prove the chosen
invariant and its bounded witness. Translation validation can reject or raise
false alarms when it cannot explain a transformation, and its validator has
non-zero cost. This experiment therefore does not claim arbitrary-program
semantic equivalence, completeness, toolchain correctness, repository writes,
or promotion authority. `OPEN` and `REFUTED` are deliberately non-authorizing.
