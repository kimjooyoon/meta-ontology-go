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

Version 1 carried six coordinates. Version 2 preserves those six coordinates
and their exact denominator map, then adds `source-semantic-causality` as a
seventh coordinate with a fixed denominator of 3. The v2 contract records the
v1 predecessor path, its digest, and the reason for this additive upgrade; it
does not silently rewrite v1.

Every candidate carries the same seven coordinates, in this order:

`source-replay`, `receipt-independence`, `counterexample-boundary`,
`unknown-localization`, `extension-evidence`, `read-only-effects`,
`source-semantic-causality`.

The denominator is fixed by `contract.json`. The numerator is the observed
integer copied from the producer receipt; it is never turned into a percentage,
estimated improvement, weighted average, rank, or winner. Each coordinate also
preserves `producer`, `consumer`, `meta-operation`, `proof-choice`, `stage`,
`step`, `reason`, and one of `OPEN`, `DISCHARGED`, or `REFUTED`.

The fixed fixture deliberately exposes different evidence shapes. The exact
v2 vectors (with `D`, `O`, and `R` meaning `DISCHARGED`, `OPEN`, and
`REFUTED`) are:

| Candidate | Coordinate vector (`numerator/denominator status`) | Counterexamples | Unknown locations |
| --- | --- | ---: | ---: |
| `derive` | `1/1 D, 1/1 D, 2/2 D, 2/2 D, 0/1 O, 1/1 D, 0/3 R` | 2 | 2 |
| `replay` | `1/1 D, 0/1 R, 1/2 O, 1/2 O, 1/1 D, 1/1 D, 0/3 R` | 1 | 1 |
| `reflect` | `1/1 D, 1/1 D, 0/2 R, 0/2 R, 0/1 R, 1/1 D, 0/3 R` | 0 | 0 |

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

## Source-semantic causality

`causality-manifest.json` is the minimum adoption schema for a source-bound
experiment: predecessor contract identity/digest, candidate source path, one
baseline operation value, one semantic intervention value, a comment-only
non-semantic intervention, and the receipt fields that must change. CI runs a
fixed three-case contract for every candidate:

`baseline`, `semantic intervention`, `non-semantic intervention`.

The source observer extracts the actual `computes` value from each `.gooo`
file. An intervention is discharged only when the semantic source value and
raw source digest move as contracted, a contracted receipt field among
`semantic_value`, `decision`, or `claim_transitions` changes, and the
non-semantic source changes only its digest while preserving the receipt's
semantic projection and decision. The independent causality report preserves
each case's exact coordinate vector and claim transitions; it emits no score,
weighted average, rank, or winner.

The current producer intentionally demonstrates the falsifier: all three
semantic interventions produce `REFUTED / DIGEST_ONLY_BINDING`, yielding
`causal_cases 0/3`, `digest_only_cases 3`, and
`hardcoded_fixture_cases 3`. The semantic source values do change, but the
receipt semantic projection, decision, and receipt claim transitions do not.
The independent audit ledger records the direct claim transitions below. All three
comment-only interventions are `DISCHARGED` because their raw source digests
change while their semantic projections and decisions remain equal. This is a
diagnostic finding about source binding, not a winner declaration.

The causality transition ledger has the fixed denominator `9` (`3 operations ×
3 intervention claims`): `REFUTED 3/9`, `DISCHARGED 6/9`, and `OPEN 0/9`.
Each baseline source-observation claim is `OPEN -> DISCHARGED`, each semantic
causality claim is `OPEN -> REFUTED / DIGEST_ONLY_BINDING`, and each
non-semantic invariance claim is `OPEN -> DISCHARGED /
SEMANTIC_PROJECTION_PRESERVED`. The `0/3` coordinate is only the count; its
status is `REFUTED` because these direct transitions are present.

`required_change_fields` is the minimum adoption contract for each case. It
must contain a non-empty subset of `semantic_value`, `decision`, and
`claim_transitions`; the consumer checks exactly the declared fields. The
current three cases declare all three, while CI also exercises a one-field
`decision` subset and rejects an empty subset.

The causality evaluator is a producer-free package and command. Actions checks
its dependency graph with `go list -deps`: forbidden producer dependencies are
`0` with an allowed maximum of `0`, and the independence contract is `1/1`.
An unknown finding is fail-closed at `LOWER_RESOLUTION`, retaining its exact
candidate, case, stage, step, and reason.

The CI workflow checks each baseline, semantic, and comment-only `.gooo`
source, produces two deterministic receipt passes for every case, observes
the source value independently, adjudicates both the portfolio and causality
reports, replays them, and mutates a receipt as a counterfactual. Local test
execution is intentionally not part of this experiment's workflow; GitHub
Actions is the verification authority.

See [the governance note](../../docs/language/language-experiment-portfolio.md)
for the external principles and the limits of this evidence.
