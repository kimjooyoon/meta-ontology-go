# Keep candidates tied to the helper that produced them

`Seed` first selects a body that doubles its input. `Build` then derives a small
set of offsets by calling that selected helper with zero and one in its hole.
The local case requires `Build(2) = 6`; an offset of one satisfies it.

Previously, caller-guided construction recreated the candidate set against the
original `Seed` body, which returned its input unchanged. The resulting plan
digest differed and stopped an otherwise valid program before caller execution.
This example reproduces that failure.

Development source now retains `body_search.initial_candidate_set` at its actual
source checkpoint. Caller-guided construction uses this fixed, bounded set.
Each `search_candidates` entry records `planning_source_sha256` separately from
the current `input_source_sha256`. Changing a helper reruns the local cases with
that helper; old scores are never reused. Saved replay verifies the original pool
against its original checkpoint and then verifies each current execution.

The finite search space is the product of these initial pools. New expressions
that would emerge from regenerating a grammar after changing a helper are outside
this experiment's pool. A partial result remains partial when the fixed pool or
program budget is exhausted. No model is configured in this example.

## Run from development source

Public compiler 0.6.29 predates this fix. Build a checkout containing retained
search candidate sets, and use a new output directory:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go build -o ./gooo ./cmd/gooo
./gooo body-construct --source examples/caller-search-checkpoint/main.gooo.fixture \
  --entry Main --attempts 8 \
  --construction-cases examples/caller-search-checkpoint/construction-cases.json \
  --cases examples/caller-search-checkpoint/evaluation-cases.json \
  --out search-checkpoint --format json > search-checkpoint.json

./gooo body-construct --source search-checkpoint/original.gooo \
  --construction search-checkpoint/construction.json \
  --cases examples/caller-search-checkpoint/evaluation-cases.json --format text
```

The native rows are finite examples. Input exposure and model-training exposure
remain separate, with `UNKNOWN` where construction coverage is incomplete.
