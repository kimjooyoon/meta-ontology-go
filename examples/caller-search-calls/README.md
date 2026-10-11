# Search an expression around an assembled helper

This development example connects a record helper to integer IR search.
Public 0.6.29 rejects this combination before caller execution. Use the source
containing the helper-call fix and Go 1.27.2 for these commands.

`Seed` first chooses its record field expression. `Divide` then calls that helper
inside a body with a denominator hole. Its source cases require `1 -> 2` and
`0 -> 1`. A candidate can call the helper successfully and still fail when it
divides by zero. Search retains that failure and continues through its declared
candidate budget. The constant denominator `1` meets the supplied cases.

From the compiler checkout:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.2 go build -o ./gooo ./cmd/gooo
./gooo body-construct --source examples/caller-search-calls/main.gooo.fixture \
  --entry Main --construction-cases examples/caller-search-calls/construction-cases.json \
  --cases examples/caller-search-calls/evaluation-cases.json \
  --attempts 5 --out search-with-helper --format json > search-with-helper.json

./gooo body-construct --source search-with-helper/original.gooo \
  --construction search-with-helper/construction.json \
  --cases examples/caller-search-calls/evaluation-cases.json --format text
```

Use a new output directory. This example uses deterministic selection and no
model. Replay verifies the source, helper implementation, failed search attempts
and selected program before running the supplied evaluation suite.

Development search receipts now include `local_calls`: helper inputs observed
during candidate-generation probes, candidate scoring, selected training and
selected holdout passes. A failed candidate keeps the calls already entered,
including earlier cases. The trace records coerced values at actual call entry;
it does not execute an expression again to collect a value.

This covers the IR search engine invocation. Separate preflight evaluations,
other assembly profiles and the combination of local and native histories still
need coverage before input separation can improve. That metric remains `UNKNOWN`
for inputs whose exposure cannot be established. See the
[local call observation contract](../../docs/local-call-observations.md).

The [saved synthetic observation](observed-local-calls-47b80a0f.json) binds a clean
development build at `47b80a0f` to the original source digest. Its one constructor
invocation retained eight helper calls across four local evaluation passes,
including `Seed(0)` before the first candidate failed. Native evaluation and
saved JSON replay each matched 3/3 supplied rows with zero model calls. Full
construction input separation remained `UNKNOWN`; these are finite observations.
