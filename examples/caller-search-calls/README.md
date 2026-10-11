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

The local evaluator still does not retain the arguments of every helper call.
Input separation therefore remains `UNKNOWN` for inputs whose exposure cannot
be established. Successful finite outputs and complete input-history coverage
are separate measurements; this change supplies the missing function connection.
