# Invariant-preserving transformation experiment

This source contract treats transformation authority as a four-part invariant:

`precondition ∧ transformation ∧ postcondition ∧ regression witness`

The producer emits a data-only receipt. The independent judge recomputes the
decision from the fixed denominator and the receipt, so a producer cannot mint
authority by changing the final decision field. Each case's `computes` value
declares an integer input, an `add:n` candidate, an expected invariant value,
replay availability, and effect policy. The producer and judge parse and
execute those values independently; a valid receipt digest alone is not
sufficient. `OPEN` means evidence is absent, `DISCHARGED` means the obligation
has a bound witness, and `REFUTED` records a counterexample. Only four fixed
cases are admitted: a preserved translation, a semantic violation, missing
regression evidence, and an approved artifact.

The approved artifact case records one separate `APPROVED_ARTIFACT_RECORDED`
effect. It still has `repository_writes=0` and `mutation_authority=false`:
recording an approved product is not permission to mutate the repository.

## Intervention witnesses

The intervention witness is a separate fixed denominator with exactly two
cases, never folded into the four-case authority coverage score:

* `semantic-change` changes the first fixture's `expected=3` to `expected=4`.
  Its parsed projection and contracted receipt change, so the independent
  judge changes `AUTHORIZED` to `REFUTED` with reason
  `SEMANTIC_POSTCONDITION_REFUTED`.
* `nonsemantic-change` appends only blank lines and a comment. Its raw
  `SourceDigest` and receipt digest change, while the parsed/lowered fixture
  projection, decision, resolution, reason, and claim transitions remain
  equal. Both source variants keep zero repository writes.

Each intervention is a persistent `OPEN -> DISCHARGED` claim with an exact
`INTERVENTION` stage, comparison step, and reason. CI publishes two separate
`1/1 (10000 BPS)` denominators and does not produce an aggregate intervention
score. The report derives `PASS` only when both claims are discharged; an
observed contradiction becomes `FAIL_CLOSED` with `REFUTED`, and an
unobservable obligation becomes `FAIL_CLOSED` with `OPEN` and lower
resolution. A report consumer checks these derived relationships without
calling the producer's `Build` function. A separately named deterministic
replay may call `Build` only to check repeatability.

## Research basis and limits

This experiment is informed by [Necula, “Translation Validation for an
Optimizing Compiler,” PLDI 2000](https://dl.acm.org/doi/10.1145/349299.349314),
which validates the actual output of each compilation and uses simulation
relations as witnesses, and [Sultana and Thompson, “Mechanical Verification of
Refactorings”](https://kar.kent.ac.uk/23959/), which formalizes refactorings in
Isabelle/HOL and treats behavior preservation as the correctness condition.

Those works do not make this small Gooo model a verified refactoring engine.
Program equivalence is undecidable; a finite receipt can only prove the chosen
invariant for four synthetic, bounded integer fixtures. Translation validation
can reject or raise false alarms when it cannot explain a transformation, and
its validator has non-zero cost. This experiment therefore does not claim
arbitrary-program semantic equivalence, completeness, toolchain correctness,
repository writes, arbitrary transformation authorization, or promotion
authority. `OPEN` and `REFUTED` are deliberately non-authorizing.
