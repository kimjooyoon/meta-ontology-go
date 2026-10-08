# Calling code can reopen a locally complete helper

The helper's local example `0 -> 0` accepts both `input` and `input * 2`.
Its caller needs `3 -> 6`. Previously, completing the helper's own example
ended that local search. The new `body-construct` command tries source-declared
combinations and immediately builds and executes each whole program. Caller
failures can therefore advance the search while the helper's obligations remain
visible. See the [runnable examples](../../../examples/caller-guided-construction).

This observation uses clean compiler source
`c43b9a15da31ac97f791a31bbfa754a8a273f790`, Go 1.27.1 and decision runtime
`v0.2.26-experimental`. That commit froze the [plan](PLAN.md) and both example
programs before model use. Its version string is still `0.6.11-dev`; this
experimental command is newer than the published 0.6.11 binary. Publication,
CI and release status must be checked separately from this local observation.

## What happened

One new three-choice program has two numeric fields and a condition. The helper
example uses zero; the caller expects `3 -> 15`. We ran that same program with
four whole-program budgets under two orders: eight runs of one program.

| Program budget | Fixed-order attempts | Fixed evaluation | Model-order attempts | Model evaluation |
| ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 1/7 | 1 | 7/7 |
| 2 | 2 | 1/7 | 1 | 7/7 |
| 4 | 4 | 1/7 | 1 | 7/7 |
| 8 | 8 | 7/7 | 1 | 7/7 |

Every attempted helper matched its one local example. Fixed order remained
`PARTIAL_FINITE` at budgets 1, 2 and 4; it satisfied the caller on its eighth
combination. Model order satisfied the caller on its first combination in each
run. The two completed routes selected byte-identical Gooo source and produced
the same seven actual evaluation results. Each mode initially checked one local
candidate; whole-program attempts and repeated local checks are separate work.

The evaluation includes one caller input already consumed during construction
and six other root inputs. Zero is also the helper's local example. These seven
rows do not establish performance on unseen tasks or a probability of satisfying
arbitrary intent.

The separate two-choice regression took two whole-program attempts. Its caller
score changed from 0/1 to 1/1 while the helper stayed at 1/1. Four evaluation rows
then passed, including an exact integer above the float64 exact-integer range.

## The model's limited job

We used the existing independently trained
[graph QAT model](https://github.com/kimjooyoon/gooo-ecosystem-workbench/tree/19485bb64278ab8ac219e042f92af677ff3f5051/models/graph-chooser-20261008/all-data-demonstration/qat_ternary).
No training or weight update occurred. Its metadata and weights SHA-256 digests
are in the frozen plan and every model observation.

The model ranked the helper's initial choices once per run. Actual caller
failures advanced the retained bounded order; they did not trigger another
prediction or update the model. The four predictions took 42.334–45.625 µs;
model setup took 0.187291–0.2605 ms and reported 2,096 resident tensor bytes.
That byte count excludes runtime and process memory.

Observed construction elapsed time was 0.597–3.130 seconds for fixed order and
0.315–0.520 seconds for model order. Budgets, build caches and sequential execution
affect these readings. They are observations of this example, not a performance
benchmark. Raw native process resource observations are retained; host CPU
utilization was not sampled.

After removing the temporary model copy, saved model construction replay
reconstructed its choices, re-executed its recorded history and reproduced all
seven evaluation results with zero new model calls. The simple regression also
replayed. Historical timings and model predictions are retained observations;
replay checks source choices and actual behavior without reattesting inference.

## Limits and additional regression coverage

The command currently handles record-choice bodies in one typed composition,
with at most 16 such bodies, 64 whole-program attempts and a three-minute search
deadline. Source attempt limits bound each eligible ranking prefix. A global
budget can combine those prefixes; it cannot expand the source choices.
Source-fill, IR-search, package-wide joint search and external Gooo policies for
joint combinations remain separate work.

Tests cover two helpers that must both change before the caller improves;
contradictory local/caller requirements that stay partial; nested helpers whose
local examples must be rechecked when dependencies change; explicit bindings;
parallel independent calls; cancellation; saved history and altered observations.
The recorded targeted race logs cover these routes and neighboring composition
and record-assembly behavior. Repository-wide CI is a separate result.

## Inspect and reproduce

`observations.tar.gz` contains eleven complete JSON command outputs, saved
construction directories for fresh runs and model replay, and compiler identity.
It contains no compiler binary or model weights. `summary.json` is derived from
the raw files by the small Go observer. It recounts actual values with exact JSON
numbers and checks fixed/model source equality and replay values.

From the repository root:

```sh
observations=$(mktemp -d)
tar -xzf docs/research/caller-guided-construction-20261008/observations.tar.gz -C "$observations"
go run ./docs/research/caller-guided-construction-20261008/observe.go "$observations"
```

To execute a retained construction with this development source:

```sh
go run ./cmd/gooo body-construct \
  --source "$observations/model-8/original.gooo" \
  --construction "$observations/model-8/construction.json" \
  --cases examples/caller-guided-construction/model-evaluation-cases.json
```

This re-executes native programs. `observe.go` only inspects the saved readings;
it does not substitute for native replay. File digests are in `checksums.sha256`.
