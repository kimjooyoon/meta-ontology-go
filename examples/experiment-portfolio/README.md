# Meta-programming experiment portfolio

This is a read-only comparison experiment for three different meta-operations.
It preserves a fixed coordinate vector per candidate instead of collapsing the
experiments into a score or declaring a winner.

The source alternatives are real Gooo declarations:

| Candidate | Gooo source | Meta-operation | Proof choice |
| --- | --- | --- | --- |
| `derive` | `alternatives/derive.gooo` | `derive-coordinate-vector` | `source-digest` |
| `replay` | `alternatives/replay.gooo` | `replay-independent-receipt` | `independent-receipt` |
| `reflect` | `alternatives/reflect.gooo` | `reflect-counterexample-boundary` | `counterexample-replay` |

Every candidate carries the same six coordinates, in this order:

`source-replay`, `receipt-independence`, `counterexample-boundary`,
`unknown-localization`, `extension-evidence`, `read-only-effects`.

The denominator is fixed by `contract.json`. The numerator is the observed
integer copied from the producer receipt; it is never turned into a percentage,
estimated improvement, weighted average, rank, or winner. Each coordinate also
preserves `producer`, `consumer`, `meta-operation`, `proof-choice`, `stage`,
`step`, `reason`, and one of `OPEN`, `DISCHARGED`, or `REFUTED`.

The fixed fixture deliberately exposes different evidence shapes:

| Candidate | Coordinate vector (`numerator/denominator status`) | Counterexamples | Unknown locations |
| --- | --- | ---: | ---: |
| `derive` | `1/1 D, 1/1 D, 2/2 D, 2/2 D, 0/1 O, 1/1 D` | 2 | 2 |
| `replay` | `1/1 D, 0/1 R, 1/2 O, 1/2 O, 1/1 D, 1/1 D` | 1 | 1 |
| `reflect` | `1/1 D, 1/1 D, 0/2 R, 0/2 R, 0/1 R, 1/1 D` | 0 | 0 |

`D`, `O`, and `R` stand for `DISCHARGED`, `OPEN`, and `REFUTED`. These are
evidence states, not a quality score. Counterexample IDs and unknown paths are
retained in their original positions. The evaluator copies every coordinate
and evidence list from each independently sealed receipt and rejects a digest
mismatch or a count mismatch.

The experiment claims only exact receipt and vector preservation for this fixed
fixture. It does not claim semantic equivalence, language quality, production
readiness, scalability, or that any candidate is better. It also does not
execute a meta-operation or estimate an improvement rate. Repository writes
must remain `0` and mutation authority must remain `false`.

The CI workflow checks the three `.gooo` sources, produces two deterministic
receipt passes, adjudicates them, replays the report, and mutates a receipt as a
counterfactual. Local test execution is intentionally not part of this
experiment's workflow; GitHub Actions is the verification authority.

See [the governance note](../../docs/language/language-experiment-portfolio.md)
for the external principles and the limits of this evidence.
