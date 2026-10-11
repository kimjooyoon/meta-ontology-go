# Search an expression around an assembled helper

This example connects a record helper to integer IR search. It requires Gooo
0.6.30 or later and Go 1.27.2. Earlier compiler 0.6.29 rejected this combination
before caller execution. See the [release guide](../../docs/releases/0.6.30-dev.md).

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

Search receipts include `local_calls`: helper inputs observed
during candidate-generation probes, candidate scoring, selected training and
selected holdout passes. A failed candidate keeps the calls already entered,
including earlier cases. The trace records coerced values at actual call entry;
it does not execute an expression again to collect a value.

The metric includes recorded local record/search inputs and consumed native
caller attempts. Gooo 0.6.30 also records search and fill preflight calls.
Construction coverage remains partial; an input stays `UNKNOWN` when exposure cannot be
established. See the
[local call observation contract](../../docs/local-call-observations.md).

The [saved synthetic observation](observed-local-calls-47b80a0f.json) binds a clean
development build at `47b80a0f` to the original source digest. Its one constructor
invocation retained eight helper calls across four local evaluation passes,
including `Seed(0)` before the first candidate failed. Native evaluation and
saved JSON replay each matched 3/3 supplied rows with zero model calls. Full
construction input separation remained `UNKNOWN`; these are finite observations.

The [later input-exposure observation](observed-input-exposure-ade411e9.json)
binds clean development source `ade411e9` to one new construction and saved
replay using the same source and case files. Including recorded `Seed(0)` now
classifies two evaluation inputs as overlapping and one as unknown. It declares
no new disjoint inputs. Native results remain 3/3, with zero model calls. The
earlier observation stays unchanged.
