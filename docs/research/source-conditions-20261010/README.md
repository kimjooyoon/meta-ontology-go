# Source conditions in the own-model construction path

Observed on 2026-10-10. Compiler producer:
`62f2252daf8f250e350db53ccea0dc6abe1d7571`, clean Go 1.27.2 build, public
decision-runtime `v0.2.27-experimental`. Compiler binary SHA-256:
`648dab6670a2a81c1e2fa0d1fba3bec9a869e00707e356fc724c51ec1f94516e`.

## What changed

The existing small Gooo model chooses a finite path label for each declared
choice. This experiment keeps its weights, local output case, caller case,
eleven evaluation inputs, and compiler identical. The second source adds only:

```gooo
condition_case "comparison" input "-1" -> "true"
condition_case "comparison" input "0" -> "false"
condition_case "comparison" input "1" -> "false"
```

The intended comparison is whether the input is negative. Both a predicate and
its branches can be reversed together while still producing a correct absolute
value. Final-output tests alone did not distinguish these structures.

The compiler now binds these expectations to source identity, checks the actual
condition during each candidate's execution, and excludes mismatching or
unreached conditions from selection. It retains their output scores. The same
rule applies when a caller tries another path, and when saved construction is
replayed. A selected body's typed structure must also match the emitted body.

## Paired observations

| Measurement | Baseline | Three source conditions |
| --- | ---: | ---: |
| Emitted predicate | `0 < input` | `input < 0` |
| Predicate observations matching `input < 0` | 1/3 | 3/3 |
| Final native caller outputs | 11/11 | 11/11 |
| Initial candidate evaluations | 1 | 3 |
| Initial candidates rejected by conditions | 0 | 2 |
| Caller construction attempts | 1 | 2 |
| Local model predictions | 3 | 3 |
| Sum of prediction time | 46,501 ns | 39,876 ns |
| Entire fresh CLI wall time | 0.84 s | 1.01 s |
| User + system CPU time | 0.52 s | 0.77 s |
| Maximum resident set reported by `time -l` | 87,687,168 B | 86,966,272 B |
| New model calls in saved replay | 0 | 0 |

One observation per arm was taken sequentially on the local Mac. CLI timing and
memory include native Go build/run subprocesses. They do not measure model-only
memory or host CPU utilization. The aggregate CPU-time/wall-time ratios are
61.9% and 76.2% of one CPU equivalent across the timed process tree; these are
not whole-machine utilization readings. The shorter prediction time is a single
observation, not a speedup claim.

The model proposed the same wrong comparison in both arms. With the conditions,
the initial finite search rejected two candidates and selected mask 0, which
satisfied the local `0 -> 0` case and all three condition cases. The caller's
`Main(3) -> 6` case then rejected that program; the next caller candidate, mask
1, satisfied both the condition contract and caller case. This illustrates why
intermediate checks and output checks remain separate dimensions.

The eleven final native inputs include `±9007199254740995`; the positive caller
output is exactly `18014398509481990`. Each saved replay reproduced all eleven
outputs without model calls. The recount uses Go's `json.Number`/`int64`, so it
does not round these values through floating point.

The baseline's 1/3 is a separate audit of its emitted predicate; its source did
not declare condition cases. The constrained arm records 3/3 during selection
and caller realization as well as the same final predicate audit. Neither count
estimates general natural-language understanding or correctness on all inputs.

## Model and reproducibility

The frozen [joint path model](https://github.com/kimjooyoon/gooo-ecosystem-workbench/tree/b63ea6d74603a89e1a74af2b3dc36c1596742935/models/joint-path-20261010)
uses 256 input features, 48 hidden units, and 8 labels. Its ternary weight file
is 2,759 bytes, SHA-256
`dcd8e44591626d421d4961bfef82ec197e947cb1d5d2cd92868d908bc7de4aed`.
No training or weight change occurred here. Earlier model accuracy regressions
remain in the original study; this result measures the compiler's use of
explicit conditions.

`original/` contains compressed, byte-preserved source, model, preflight, native
construction, generated Go, runtime/replay, timing, and producer records.
`SHA256SUMS` covers the compressed files. `recount.json` is derived from them:

```sh
cd docs/research/source-conditions-20261010
shasum -a 256 -c SHA256SUMS
GOWORK=off GOTOOLCHAIN=go1.27.2 go run recount.go original > /tmp/source-conditions-recount.json
cmp recount.json /tmp/source-conditions-recount.json
```

To make a fresh run, decompress the sources/model/case files into a new
directory. From there, run a compiler built from the producer commit:

```sh
gooo body-construct --source constrained.gooo --entry Main \
  --construction-cases construction-cases.json --cases evaluation-cases.json \
  --attempts 8 --model model/model.json --out fresh-constrained
gooo body-construct --source fresh-constrained/original.gooo \
  --construction fresh-constrained/construction.json --cases evaluation-cases.json
```

The broader [language fixture](../../../examples/body-codegen/source-condition-cases.gooo.fixture)
also covers exact large-int condition inputs and seven local output cases.
Its selection evidence is distinct from this paired experiment's single local
output case. See [the grammar and limits](../../source-assembly.md#intermediate-conditions).

## What remains

The present decision model still ranks paths whose intermediate behavior can
conflict with a source hint. A subsequent training experiment should derive
compatible target paths from both explicit output and condition cases, retain
multiple compatible paths, and measure unseen Korean/English phrasing
separately. Source expectations cannot be inferred from a model score alone.
This implementation currently covers finite `if` conditions in Integer →
Integer typed paths; record-field predicates and other intermediate values
need their own source semantics and execution evidence.
