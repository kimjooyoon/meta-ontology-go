# Compose generated Gooo bodies through declared binds

`body-compose` constructs the activities in one Gooo file, combines their checked
Go functions, and immediately builds and executes the declared graph. A selected
integer assembly can feed another integer activity, an Integer -> Boolean body,
a Boolean -> Boolean body, and fan out to several consumers. Independent Text
roots and Text -> Boolean edges use the same typed graph.

The source is the assembly drawing and the `bind` declarations are its wires.
Each root gets an explicitly named input. Bound activities get only their
declared producer's result. The runtime retains every activity's actual input
and output, including intermediate results that have no supplied expectation.

## Run and continue

```sh
gooo body-compose \
  --source examples/body-codegen/native-composition.gooo.fixture \
  --cases examples/body-codegen/native-composition-cases.json \
  --out /tmp/gooo-composition
```

Use a new output directory. Optional `--model /path/to/model.json` loads one
local path model for all `assembling` activities in this request. Without it,
assembly uses deterministic bounded search. Ordinary bodies are preserved and
typechecked. The current assembly model profile remains Integer -> Integer;
Boolean and Text bodies connect through ordinary checked source lowering.

The directory contains `original.gooo`, `cases.json`, `composition.json`,
`runtime.json`, `realized.gooo`, `generated.go`, `main.go` and `go.mod`.
The final three files form the actual runnable program; `main.go` contains the
generated input delivery calls. Selected source checkpoints are retained in
`realized.gooo` and can be the source of another composition request.

Execute a saved composition with a new finite suite:

```sh
gooo body-compose \
  --source /tmp/gooo-composition/original.gooo \
  --composition /tmp/gooo-composition/composition.json \
  --cases /tmp/gooo-composition/cases.json
```

Replay reconstructs every step and both Go files before compiling. It loads no
model and makes zero predictions. Each assembly step retains the exact preceding
Gooo source digest; its checkpoint becomes the next activity's source. The saved
generation records preserve original model and finite selection observations.

## Cases and completeness

```json
{
  "schema": "gooo/body-composition-cases/v1",
  "cases": [{
    "inputs": {"Assemble": -8, "Decorate": "gooo"},
    "expected": {"Clamp": 10, "Invert": false, "Same": true}
  }]
}
```

Each case supplies exactly every unbound activity's input and at least one named
expectation. Expectations may cover any declared activity, including a root or
an intermediate stage. Integer values are exact int64 JSON integers, Boolean
values are JSON booleans, and Text values are strings. Missing values and null
are distinct from zero, false and the empty string.

`finite_passed / finite_total` counts these supplied runtime expectations. An
unobserved stage output is retained without receiving an accuracy credit.
Partial finite results remain visible even when both compiled executions agree.
The generation-time selection examples and the runtime suite are separate;
the same input may occur in both, so no holdout accuracy is inferred.

## Execution and layout

The existing typed-plan compiler checks exact entity identities, ports and
cycles. Activity order is its canonical topological order; edges retain stable
producer/consumer/entity IDs. A consumer has one producer and a producer may
have many consumers. Multiple roots require separate explicit inputs. There
are no inferred edges or runtime dependency waits.

Graph preparation uses a 16-entry activity array. The generated driver uses
fixed-width arrays for root inputs and ordered outputs, with typed local values
between calls. It allocates output JSON storage per case, rather than a map at
each graph step. Generation still retains complete per-activity receipts; those
records and text values have separate memory costs.

One native build is followed immediately by two executions of the entire graph.
Existing Go tool selection, bounded child processes and process resource
observations are reused. Each request owns a fresh workspace, removed when the
request finishes; this first composition path has no executable cache. Native
execution records build/run time, CPU time and process peak RSS. These values
do not measure whole-computer CPU utilization or prove a speed gain.

## Current boundaries

The graph supports 2..16 activities with one input and one result each, explicit
in-invocation binds, and the ordinary pure Integer/Boolean/Text body profile.
Source is bounded to128 KiB, generated function source to256 KiB, suites to32 KiB
and1..128 cases, and each decoded Text scalar to1024 UTF-8 bytes. Child stdout
and stderr retain the existing64 KiB limit. A total request has a60-second native
budget and each native run a2-second budget; cancellations and failures retain
their first stage. Feedback across invocations, joins with multiple producers,
record-valued bodies, calls and loops require further language work.

Related: [source assembly](source-assembly.md), [language direction](language-direction.ko.md),
[body generation](language/body-codegen.md).
