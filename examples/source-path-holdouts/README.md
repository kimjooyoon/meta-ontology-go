# Choose with one set of examples, measure with another

`Fee` starts with two permitted subtraction directions. Both match the source
case `47 -> 0`. Its `holdout_case "59" -> "-12"` describes a separate evaluation;
that expected answer is excluded from path ranking, model inputs and observation
questions.

The initial deterministic body is `input - 47`. It matches the selection case
and misses the holdout. The caller's construction example then requires
`Main(72) = -30`. Since Main calls `Fee(input + 5)`, caller-guided construction
selects `47 - input`. The final body is measured on the holdout after selection.

```sh
GOWORK=off GOTOOLCHAIN=go1.27.3 go run ./cmd/gooo body-construct \
  --source examples/source-path-holdouts/main.gooo.fixture --entry Main \
  --construction-cases examples/source-path-holdouts/construction.json \
  --cases examples/source-path-holdouts/evaluation.json --attempts 2 \
  --format text --out /tmp/gooo-path-holdout-example
```

Expected observations:

| Observation | Expected result |
| --- | --- |
| Initial local selection | One source case matches; holdout 0/1 |
| Caller-guided selection | Second of two candidates; local 1/1, caller 1/1 |
| Final selected body | Holdout 1/1; its input was absent from caller attempts |
| Separate native evaluation | Three supplied outputs match |

Changing the holdout expected value to `12` reverses its score and leaves the
chosen candidate and stopping decision unchanged. A held mismatch is retained;
it does not start another selection attempt. A bounded evaluator error leaves
the accuracy unavailable and retains its cause.

## A root input can hide an overlapping helper input

Use `overlapping-construction.json` instead of `construction.json` and choose a
different output directory. The caller input is 54, but Fee receives 59.
The final report therefore lists 59 as observed during caller selection. It
does not describe this holdout as independent of that feedback.

The overlap observation follows actual activity arguments in every consumed
native caller attempt, including unsuccessful candidates. Other local helper
evaluations and prior model-training exposure are outside that measure. Missing
caller observations remain unknown.

The source contract and the initial/final evaluation records survive saved
construction replay. `body-construct --report <saved-result.json>` only renders
the saved record; it does not execute a program or invoke a model.

This is a development feature for Integer typed paths. It measures the named
finite cases and keeps the selection/evaluation distinction visible. Public
compiler 0.6.34 does not include it.
