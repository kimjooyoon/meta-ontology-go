# Observe inputs used before candidate selection

This development example records the pure probes used to validate an IR-search
grammar before choosing a body. Public compiler 0.6.29 predates this feature.

`Build` must choose an offset inside `Step(input + offset)`. Its source case is
`2 -> 6`, and `Step` doubles its argument. The residual grammar probes offsets
zero and one, entering `Step(2)` and `Step(3)`. Those inputs were used even before
the constructor selected a candidate. The constructor retains the actual calls
at both its initial preflight and the later activity-generation checkpoint.

From a compiler checkout containing this feature, with Go 1.27.2:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go build -o ./gooo ./cmd/gooo
./gooo body-construct --source examples/caller-search-validation/main.gooo.fixture \
  --entry Main --construction-cases examples/caller-search-validation/construction-cases.json \
  --cases examples/caller-search-validation/evaluation-cases.json \
  --attempts 5 --out search-validation --format json > search-validation.json

./gooo body-construct --source search-validation/original.gooo \
  --construction search-validation/construction.json \
  --cases examples/caller-search-validation/evaluation-cases.json --format text
```

Use a new output directory. This example uses deterministic selection and makes
no model calls. Inspect `initial.search_validations` in the saved construction:
each entry contains a stage, source digest, activity identity and original
`local_calls` history. A validation may complete despite one probe failing;
each probe keeps its own completion status and entered calls.

Saved replay checks the recorded boundaries against their exact source
checkpoints. It repeats pure validation probes and does not select another
candidate. Historical receipts with no such history remain readable.

This expands the evidence available to the input-exposure metric. Remaining
unrecorded construction boundaries still prevent a complete independence claim.
The three evaluation rows are a finite example, not an estimate of general
correctness. See the [observation contract](../../docs/local-call-observations.md).
