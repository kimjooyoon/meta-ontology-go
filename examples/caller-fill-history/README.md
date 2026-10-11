# Inputs used by a rejected fill candidate

This development example assembles `Seed`, then fills two expressions in `Build`.
It shows why failed local evaluations belong to a program's construction history.

The first assignment calls `Seed(2)` and then `Seed(0)`. Its second case divides
by zero after entering the helper. Gooo keeps that prefix, rejects the assignment
without a completed accuracy score, and evaluates the next declared assignment.
The selected assignment matches both local cases and its separate holdout case.

`Main(6)` later reaches `Seed(0)`, while `Main(8)` reaches `Seed(2)`. Both root
inputs are different from the local training inputs, but the inner function has
already received these values. The input-separation report counts two overlapping
inputs and leaves the remaining input unknown. Fill preflight and other construction
boundaries still have incomplete capture. This example establishes no general
correctness or model-quality result.

## Run from development source

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
