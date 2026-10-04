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

Activities may receive 1..16 scalar inputs. One input keeps the existing `input`
name. Multiple inputs use `input0`, `input1`, and so on in declaration order,
including repeated types:

```gooo
activity Add(Integer, Integer) -> Integer computes "return input0 + input1"
bind Left.result -> Add.input0
bind Right.result -> Add.input1
```

The two inputs remain separate slots even though PROV records one unique
`used Integer` relation. Source positions, entity identities and order travel
through the bidirectional model and semantic IR; the Go signature uses that order.

Run the repeated-input, mixed-type and partly bound example:

```sh
gooo body-compose \
  --source examples/body-codegen/native-input-joins.gooo.fixture \
  --cases examples/body-codegen/native-input-joins-cases.json \
  --out /tmp/gooo-input-joins
```

Optional model ranking can assemble `Left` before its result joins `Right`.
The Add, Difference, Label and Compare bodies lower their declared source
deterministically. The finite suite has seven input scenarios and 49 named
output expectations, covering operand order, zero/false/empty values, Unicode
and int64 wraparound.

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

Each case supplies exactly every unbound input and at least one named
expectation. Expectations may cover any declared activity, including a root or
an intermediate stage. Integer values are exact int64 JSON integers, Boolean
values are JSON booleans, and Text values are strings. Missing values and null
are distinct from zero, false and the empty string.

For a multiple-input root, use keys such as `Compare.input0` and `Compare.input1`.
For a partly bound activity, supply only its unbound ports: `Label.input1` and
`Label.input2` when `Label.input0` receives Add's output. A single-input root
keeps its activity name as the case key. Supplied values cannot override a bind.

Multiple-input plan entries include an ordered `inputs` list with each port's
type, entity ID and producer index (`-1` for external input). The legacy scalar
input fields describe the first port. Runtime entries use a matching `inputs`
list with the actual value, entity ID and producer ID for each port; these
observations add no score points beyond the named expected outputs.

`finite_passed / finite_total` counts these supplied runtime expectations. An
unobserved stage output is retained without receiving an accuracy credit.
Partial finite results remain visible even when both compiled executions agree.
The generation-time selection examples and the runtime suite are separate;
the same input may occur in both, so no holdout accuracy is inferred.

## Execution and layout

The existing typed-plan compiler checks exact entity identities, ports and
cycles. Activity order is its canonical topological order; edges retain stable
producer/consumer/entity IDs. Each input port has at most one producer and a producer may
have many consumers. Multiple roots require separate explicit inputs. There
are no inferred edges or runtime dependency waits.

Graph preparation uses a 16-entry activity array. The generated driver uses
fixed-width arrays for root inputs and ordered outputs, with typed local values
between calls. It allocates output JSON storage per case, rather than a map at
each graph step. Generation still retains complete per-activity receipts; those
records and text values have separate memory costs.
The bounded input-row workspace holds at most 256 external slots (16 activities
times 16 inputs); generated call arguments follow each source signature.

One native build is followed immediately by two executions of the entire graph.
Existing Go tool selection, bounded child processes and process resource
observations are reused. Each request owns a fresh workspace, removed when the
request finishes; this first composition path has no executable cache. Native
execution records build/run time, CPU time and process peak RSS. These values
do not measure whole-computer CPU utilization or prove a speed gain.

## Current boundaries

The graph supports 2..16 activities with 1..16 inputs and one result each, explicit
in-invocation binds, and the ordinary pure Integer/Boolean/Text body profile.
Source is bounded to128 KiB, generated function source to256 KiB, suites to32 KiB
and1..128 cases, and each decoded Text scalar to1024 UTF-8 bytes. Child stdout
and stderr retain the existing64 KiB limit. A total request has a60-second native
budget and each native run a2-second budget; cancellations and failures retain
their first stage. Optional learned source assembly currently accepts a single
Integer input and Integer result; multiple-input bodies connect these selected
results through checked ordinary source. The `run` command's registered operation
profile has its own input restrictions. Feedback across invocations, record-valued
bodies, calls and loops require further language work.

Related: [source assembly](source-assembly.md), [language direction](language-direction.ko.md),
[body generation](language/body-codegen.md).
