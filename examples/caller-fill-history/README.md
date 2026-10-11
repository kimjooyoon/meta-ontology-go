# Inputs used by a rejected fill candidate

This example requires Gooo 0.6.30 or later. It assembles `Seed`, then fills two expressions in `Build`.
It shows why failed local evaluations belong to a program's construction history.

The first assignment calls `Seed(2)` and then `Seed(0)`. Its second case divides
by zero after entering the helper. Gooo keeps that prefix, rejects the assignment
without a completed accuracy score, and evaluates the next declared assignment.
The selected assignment matches both local cases and its separate holdout case.

`Main(6)` later reaches `Seed(0)`, while `Main(8)` reaches `Seed(2)`. Both root
inputs are different from the local training inputs, but the inner function has
already received these values. The input-separation report counts two overlapping
inputs and leaves the remaining input unknown. The recorded producer predates
fill-preflight capture. Gooo 0.6.30 records that boundary, while
other construction coverage remains incomplete. This example establishes no general
correctness or model-quality result.

## Run from a compiler checkout

Use a checkout containing fill-local observation support and a new output directory:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go build -o ./gooo ./cmd/gooo
./gooo body-construct --source examples/caller-fill-history/main.gooo.fixture \
  --entry Main --attempts 4 \
  --construction-cases examples/caller-fill-history/construction-cases.json \
  --cases examples/caller-fill-history/evaluation-cases.json \
  --out fill-history --format json > fill-history.json

./gooo body-construct --source fill-history/original.gooo \
  --construction fill-history/construction.json \
  --cases examples/caller-fill-history/evaluation-cases.json --format text
```

Both selection and saved replay are deterministic here; no model is configured.
The saved construction binds helper identities, source digests, candidate IDs,
pass order and typed input tuples. Replay checks those observations during its
existing evaluator passes. Changing an entered helper argument invalidates replay.

The regression test checks all three final outputs, saved JSON replay with zero
model calls, two overlapping inputs and one unknown input. A finite `3/3` output
result counts the supplied rows; those rows are not three independent observations.

## Saved CLI observation

The [source-bound observation](observed-local-calls-47cb0221.json) uses clean
development producer `47cb0221f52472504b5278edfe43706c8b303bdc`, Go 1.27.2 and
decision runtime 0.2.41. One constructor/evaluation command and one saved replay
each matched all three native rows with zero model calls. Both reports retained
two overlapping inputs and one unknown input. The failed first assignment keeps
its two entered helper calls; the successful assignment retains training, probe
and selected-case passes. Caller-selected assignments have separate histories.

Whole-command accounting, including compilation and subprocesses, measured
0.64 seconds wall, 0.29 seconds user CPU, 0.23 seconds system CPU and
87,080,960 bytes maximum RSS. Saved replay measured 0.35 seconds wall,
0.17 seconds user CPU, 0.11 seconds system CPU and 86,376,448 bytes maximum RSS.
These are single observations. Instantaneous CPU utilization and a controlled
performance comparison were not measured. This observation belongs to the
development source above. The [0.6.30 release guide](../../docs/releases/0.6.30-dev.md)
links the separately produced native platform evidence.
