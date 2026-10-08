# Integer division in a constructed Gooo body

`Divide` computes a quotient and remainder from two int64 inputs. The Gooo body
sets a validity flag, guards a zero divisor, and assigns both results to local
variables. Three declared field choices assemble the returned record. The
construction expectations belong to the source; eight separate execution tuples
cover signs, zero, and the limits of int64.

From this compiler repository:

```sh
go run ./cmd/gooo body-compose --source examples/integer-division/source.gooo.fixture \
  --cases examples/integer-division/cases.json --out out/integer-division

go run ./cmd/gooo body-compose --source out/integer-division/original.gooo \
  --composition out/integer-division/composition.json \
  --cases examples/integer-division/cases.json
```

The first command uses fixed candidate ordering. Add `--model /path/to/model.json`
to use an existing three-choice record model. Saved execution loads no model.

Quotients truncate toward zero: `-13 / 5` gives `-2`; `-13 % 5` gives `-3`.
Runtime `MinInt64 / -1` follows Go and returns `MinInt64`, with a zero remainder.
For an input divisor of zero this particular source returns `valid=false` and
two zero values. Other programs can declare a different guarded behavior.

The body evaluator returns a zero-divisor error when a guard is absent. The
native generated program follows Go's runtime failure for the same expression.
The tests compare the evaluator with separately compiled generated Go, check
short-circuit guards, and execute/replay this record program. The bounded
observations concern the supplied input tuples.
