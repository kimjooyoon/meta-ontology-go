# A local model ranks a helper, then Gooo fills its caller

This development example joins two construction stages in one Gooo source.
`Choose` declares variable assignments, a comparison, alternative branch layouts
and output/condition cases. The public source-relative model can rank those
declared paths. Gooo evaluates them within the source attempt budget.

`Build` then calls the assembled `Choose` function and picks a complete assignment
for its two expression holes. Its finite cases require adding one. This second
stage uses deterministic scoring. Saved execution reuses the assembled source.

Use the model file and checksum from the
[source-relative model guide](../../docs/source-relative-contract-model.md), then
set `model` to that local file. From a checkout containing helper-fill support:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go build -o ./gooo ./cmd/gooo
./gooo body-construct --source examples/caller-model-fill/main.gooo.fixture \
  --entry Main --model "$model" --attempts 4 \
  --construction-cases examples/caller-model-fill/construction-cases.json \
  --cases examples/caller-model-fill/evaluation-cases.json \
  --out model-and-fill --format json > model-and-fill.json

./gooo body-construct --source model-and-fill/original.gooo \
  --construction model-and-fill/construction.json \
  --cases examples/caller-model-fill/evaluation-cases.json --format text
```

Use a new output directory. Omitting `--model` uses deterministic ordering for
`Choose` too. An unreadable or incompatible model file returns an error. The
source-relative model uses `--model` for source choices; the separate
`--fill-model` option expects the older operation-classifier format and does not
accept this research model.

The model ranks a bounded set of source-owned decisions. Gooo retains the cases,
candidate scores and actual selection. A correct final program can follow a
weak model proposal, so its finite score must be read separately from the model
ranking. The source includes Korean and English intent text plus typed cases;
one such example cannot measure Korean or English language understanding.

The original observation below predates local helper-input recording during fill
preflight and selection. Newer development source records those boundaries;
remaining construction coverage is still incomplete. Input separation for this
pipeline can remain `UNKNOWN`, even when the supplied native cases pass.

## Recorded use of the public model

The [saved observation](observed-model-fill-deb93c0a.json) uses clean development
producer `deb93c0a` and the unchanged public source-relative model with SHA256
`cae32bdad17ab818498c6099066918cd9c9c591eb118afe1bc6a7ef82dc83cc4`.
One model prediction proposed mask 0, which matched 0/4 declared output cases.
The next candidate, mask 1, matched 4/4. Gooo selected that source, then the
deterministic fill selected `increment` with 1/1 training and 1/1 holdout matches.
Native evaluation and saved replay each passed 3/3 rows. Replay made zero model
calls. All three evaluation inputs remained `UNKNOWN` for construction exposure.

The model reported 76,136 resident tensor bytes, 6.376458 ms setup and 93,083 ns
for ranking. Whole-command accounting including compilation and subprocesses
reported 0.91 seconds wall, 0.32 seconds user CPU, 0.23 seconds system CPU and
87,015,424 bytes maximum RSS. Saved replay reported 0.36 seconds wall,
0.18 seconds user CPU, 0.11 seconds system CPU and 87,179,264 bytes maximum RSS.

This single observation demonstrates the connection and retains the weak first
proposal. A matched deterministic comparison was not run, so the observation
does not establish a search or speed advantage. Model training and the original
published studies were not repeated.
