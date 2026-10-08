# Compose generated Gooo bodies through declared binds

`body-compose` constructs the activities in one Gooo file, combines their checked
Go functions, and immediately builds and executes the declared graph. A single
activity runs directly without a synthetic `bind`; graphs with multiple
activities still require explicit typed binds for every data-flow edge. A selected
integer assembly can feed another integer activity, an Integer -> Boolean body,
a Boolean -> Boolean body, and fan out to several consumers. Independent Text
roots and Text -> Boolean edges use the same typed graph.

The source is the assembly drawing and the `bind` declarations are its wires.
Each root gets an explicitly named input. Bound activities get only their
declared producer's result. The runtime retains every activity's actual input
and output, including intermediate results that have no supplied expectation.

Activities may receive 1..16 scalar or declared record inputs. One input keeps the existing `input`
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

Use a new output directory. Optional `--model /path/to/model.json` retains a
local structural model for choice-based integer or record assembly. Optional
`--fill-model /path/to/model.json` retains a TinyGo operation model for
`source_fill` assignments. These are separate model profiles; a mixed graph can
use both. Empty model paths use deterministic bounded selection. All ordinary
bodies and declared assignments are checked before an optional model is loaded.

The directory contains `original.gooo`, `cases.json`, `composition.json`,
`runtime.json`, `realized.gooo`, `generated.go`, `main.go` and `go.mod`.
The final three files form the actual runnable program; `main.go` contains the
generated input delivery calls. Selected source checkpoints are retained in
`realized.gooo` and can be the source of another composition request.

## Source IR search across activities

A source `search hole` contract can now participate in the same explicit-bind
graph. Each activity enumerates its declared integer grammar, uses its own
attempt budget and finite selection cases, then emits a checked body. Holdout
cases are observed after selection. The selected expression and finite evidence
are replayed before the completed source is handed to the next activity.

```sh
gooo body-compose \
  --source examples/body-codegen/source-search-composition.gooo.fixture \
  --cases examples/body-codegen/source-search-composition-cases.json \
  --out /tmp/gooo-source-search-composition
```

The example connects negative-value normalization, an offset search and a
Boolean check. Five runtime inputs carry 15 named expectations, including
integers above JavaScript's exact-number range. `realized.gooo` contains the
selected ordinary bodies; their consumed `search` blocks are removed. Original
contracts and selection/holdout receipts stay in `original.gooo` and
`composition.json`. A saved composition reconstructs every checkpoint and
checks the selected expressions against the original candidate sets before
executing, with zero model calls.

For a single saved source-search generation, `body-realize --source ...
--generation ... --out <new-directory>` produces the same kind of reusable checkpoint.

This composition route uses deterministic ordering for source IR search.
`--model` applies to choice-based scalar or record assembly in a mixed graph;
a graph containing only source IR search rejects that option before loading a
model. No provider is inferred from environment variables in `body-compose`.
Multi-hole `source_fill` plans use the additional route below.

## Source-owned multi-hole construction

```sh
gooo body-compose \
  --source examples/body-codegen/source-fill-composition.gooo.fixture \
  --cases examples/body-codegen/source-fill-composition-cases.json \
  --out /tmp/gooo-source-fill-composition
```

This example assembles two activities with two holes each, connects their outputs
and checks a Boolean result: add one, multiply by two, then test positivity.
Five runtime inputs carry 15 named output expectations. An input above 2^53
checks that the native integer path preserves exact values.

Add `--fill-model /path/to/model.json` to use a compatible local TinyGo operation
bundle. It loads once for the graph, then predicts once for each fill activity.
The model reads the source intent and proposes an operation; Gooo maps that to
a declared complete assignment and retains both proposed and selected scores.
If a proposal scores below the best declared selection-suite candidate, Gooo
uses that candidate and records the adjustment. Record-valued fills, including
grammar-derived assignments, can also feed typed downstream activities.

`fill_model` records the one-time load. Per-activity `body_fill` receipts contain
model digests, inference timings, filled expressions, finite selection scores
and post-selection holdouts. The two-stage example also runs with a synthetic
constant-operation test model: an intentionally wrong second proposal is kept
in the receipt beside the corrected selection. This checks integration; it does
not measure a trained model's general accuracy.

Saved `--composition` replay requires no model file or provider. It reconstructs
the original decision request, recomputes every candidate score and the selected
case/holdout observations, and compares the emitted bytes before native execution.
The historical decision is reused without fresh inference. The completed
`realized.gooo` has ordinary bodies and can be composed again. `body-realize`
also accepts a saved source-fill generation. Receipts from older versions that
lack `original_source_digest` must be regenerated for this route.

The bounded deterministic scorer also serves as preflight, so generation and
replay repeat candidate evaluation. This is currently a reproducible construction
path, with no measured speed advantage. Finite scores describe only the declared
cases; incomplete grammar exploration and unmatched cases remain visible.

### Trained-model observation

The [2026-10-07 paired observation](research/source-fill-dogfood-20261007.json)
used the independently trained QAT operation model on this example. Its two
operation predictions were `equal` and `add` for the expected `add` and
`multiply`: 0/2 matched. Gooo's finite-case selection still produced 15/15 native
outputs in both the model and deterministic arms. Each matched both unique
runtime cases disjoint from construction observations. Model-training exposure
is unknown. The model added no demonstrated benefit on this workload.

Model setup took 0.165 ms and the two decision calls took 28.125 and 24.958 us.
The model files totaled 3,999 bytes; they were deleted before a successful saved
replay with zero new model calls. Whole-command user/system CPU time was
0.14/0.10 seconds in each arm. Maximum RSS, including the native build/execution
command observation, was about 81.7/82.9 MiB; host CPU utilization was not sampled.
One model-first pair is too small and cache-sensitive to establish a resource
improvement. [Model](research/source-fill-native-20261007/model-native.json),
[deterministic](research/source-fill-native-20261007/deterministic-native.json) and
[saved replay](research/source-fill-native-20261007/saved-replay.json) receipts
retain the incorrect predictions and the corrected outputs together.

## One activity without a synthetic bind

A single activity is a complete executable plan by itself. Its case keys use
the activity name, and the same command generates the projection, builds it,
and executes it twice without inventing an edge:

```sh
go run ./cmd/gooo body-compose \
  --source examples/body-codegen/native-single-activity.gooo.fixture \
  --cases examples/body-codegen/native-single-activity-cases.json \
  --go-bin "$(go env GOROOT)/bin/go"
```

The example has a local variable and a three-way conditional. The native
runtime receipt reports the finite result as `3/3` and records both runs and
their resource observations. Multiple activities still need explicit typed
binds for every internal data-flow edge.

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
The generation-time examples and runtime suite may contain the same inputs.
The following measurement exposes that overlap.

### Input separation

After source replay and two matching native executions, `runtime.json` includes
`input_separation`. It uses unique, typed root-input tuples as whole cases:

- `unique_inputs` counts distinct root inputs; `duplicate_rows` counts repeats.
- `overlapping_inputs` counts cases whose actual input at any assembling
  activity occurs in source selection/holdout cases or recorded probes.
- `disjoint_inputs` counts cases with new actual inputs at every assembling
  activity; downstream inputs come from the native delivery trace.
- `disjoint_cases_passed` counts disjoint cases satisfying every supplied
  expectation, including expectations from duplicate rows.
- `unknown_inputs` retains cases whose input separation cannot be established.

The fraction is `disjoint_cases_passed / disjoint_inputs`. One case may name
several outputs, so its unit differs from `finite_passed / finite_total`.
Optional absence remains distinct from zero, record key order is normalized,
and int64 values retain their exact identity.

With complete input observations, a positive denominator is `PASS` when all
disjoint cases match and `PROGRESS` otherwise, including measured `0/N`.
No disjoint cases, no assembling activity, unexecuted runtime, or unavailable
input identities produce `UNKNOWN` with a reason. Recorded holdouts are included
conservatively even when they were excluded from candidate ranking. External
feedback without input identities also leaves the measurement `UNKNOWN`.
Model-training exposure remains unknown; this finite fraction describes only
the recorded input boundary. The metric is descriptive and adds no CI gate.

Called-body construction also records actual helper arguments at native function
entry in `traces[].calls`, with the helper and enclosing graph activity IDs.
`called_inputs_observed` distinguishes a completed observation with no executed
call from a missing observation. New root values can therefore be classified as
overlapping when the caller maps them to a previously seen helper input. Repeated
and nested calls retain their execution order; every executed constructed input
must be disjoint for a case to receive disjoint-input credit. A path that executes
no constructed activity remains unknown.

The runtime derives an instrumented projection and driver after replaying the
saved construction. Their separate `observed_projection_sha256` and
`observed_driver_sha256` bind that observation. The saved pure projection and its
digest are preserved. Both native runs must produce identical outputs and call
observations. The retained executor also keys its artifact by these observation
digests and observes current arguments on each invocation.

Calls made while scoring another assembling body's candidates are a remaining
boundary. Their argument histories are not yet recorded. If such dependencies
exist, an apparent new input remains `UNKNOWN` with
`CONSTRUCTION_CALL_INPUTS_NOT_OBSERVED`; a directly recorded overlap can still be
identified. This preserves the difference between runtime observations and the
input history used during construction. The call recorder bounds observations
to 65,536 per root case, matching 16 graph activities with at most 4,096 calls
each; the existing execution time and output-size limits also apply.

Workspace execution also includes body fills completed before composition.
The earlier fill's selected body is matched to the executed activity, and its
training, holdout and probe tuples join the known-input sets. A new root can
still overlap when an intermediate activity receives a previously observed
value. `earlier_stages` records the activity, source, plan and canonical input-set
digests for these stages. Source-owned fills and external fill plans use the
same typed tuple accounting. See the [workspace input example](../examples/workspace-input-observations/README.md).

## Execution and layout

### Rejected body-fill assignments

`body-construct` can continue past type errors and local training-evaluation
errors in `source_fill` assignments. The initial fill receipt separates scored
candidates from `rejected_candidates`. The original assignment list remains the
search space, and rejected combinations consume the same program-attempt budget.
Their native runtime is empty; local totals cover only an already scored prefix.
`gooo/joint-construction/v5` binds these observations and their saved replay.
An initial plan with no valid assignment stops before model loading and exposes
`initial.fill_failure`, including every rejection and the plan digest.

The [five-assignment example](../examples/caller-fill-rejection/README.md)
demonstrates these cases. Malformed contracts, cancellation, holdout-evaluation
errors and native build/run errors still stop the 0.6.15 request. Holdout observations
remain separate from selection scores.

### Native arithmetic outcomes in development source

Newer development source derives a separately hashed execution observation from
the replayed pure projection. Reached int64 division or remainder by zero records
the original activity, expression span/hash and exact evaluated operands.
An affected graph activity has no actual result; dependent activities carry
`blocked_by` producer IDs. Earlier and independent deliveries and later rows
continue. Both fresh executions must emit identical outcomes.

`outcomes` counts supplied expectations as `matched`, `mismatched`, `faulted`,
`blocked` or `unobserved`. A complete observation has zero unobserved expectations;
an incomplete process observation does not supply this aggregate. A native
language fault uses `body-composition-runtime/v3`; successful historical shapes
remain readable. Fault-containing construction histories use
`joint-construction/v6`, consume their original program budget and retain local
scores separately. Saved replay reconstructs each attempted combination without
model calls. The [six-assignment example](../examples/caller-native-failure/README.md)
retains the original 0.6.15 failure and all original expected values.

Toolchain errors, unknown process faults, timeouts and cancellation remain
terminal. The generated standalone program and saved pure projection retain
their original arithmetic behavior; the execution observation has its own hashes.

### Native execution

The existing typed-plan compiler checks exact entity identities, ports and
cycles. A singleton plan has one declared activity and an empty edge list;
multiple activities without an explicit edge are rejected. Activity order is
its canonical topological order; edges retain stable producer/consumer/entity
IDs. Each input port has at most one producer and a producer may have many
consumers. Multiple roots require separate explicit inputs. There are no
inferred edges or runtime dependency waits.

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

## Contextual hole construction

The [Gooo fixture](../examples/body-codegen/hole-context-composition.gooo.fixture)
connects three activities whose missing expressions occur inside existing bodies:
`input + hole`, `(input + hole) * 2`, and a local-variable computation.
Their source-owned `integer-hole-residual/v1` grammar constructs `1`, `3`, and
`input * 4 + 3` from the declared training cases.

```sh
go run ./cmd/gooo body-compose \
  --source examples/body-codegen/hole-context-composition.gooo.fixture \
  --cases examples/body-codegen/hole-context-composition-cases.json
```

Run with a Go 1.27.1 toolchain; pass `--go-bin` when the Go binary on PATH differs.
The native suite contains four input cases and twelve output expectations. Three
inputs overlap construction observations; one input beyond 2^53 is disjoint at
all three activities. Report these denominators separately. A generated search
receipt records four pure probe evaluations per activity; composition preflight
and receipt replay can repeat that work. The native runtime makes zero model
calls. This example exercises a deterministic language extension; it does not
measure a learned model's prediction accuracy.

The [recorded comparison](research/hole-context-20261007/summary.json) at compiler
`44b16c3e` observes 0/12 native expectations with the legacy grammar and 12/12
with the contextual grammar; saved replay also observes 12/12. The change alters
the candidate space. One sequential pair on a shared host establishes this
example's construction behavior; it does not establish a general speed or
model-quality advantage. Raw receipts retain the unsuccessful baseline.

## Reconsider integer IR expressions with caller feedback

The development `body-construct` path also accepts source-declared integer
`search hole` bodies alongside record choices. Its source grammar derives
expressions; the caller's native execution can select a different expression
even when the initial local examples already pass. Each activity's `attempts`
still bounds its eligible prefix, separately from the whole-program budget.

The [caller IR search example](../examples/caller-ir-search/README.md) covers
scalar-only and mixed construction. Mixed receipts use
`gooo/joint-construction/v2`: `candidate_kinds` distinguishes record masks from
source-search indices, and `search_candidates` retains exact local observations
and grammar coverage. Record-only receipts keep v1. Replay checks both formats
against the source and re-executes caller cases with no new inference. The
optional graph model orders record choices; integer expression ordering remains
deterministic. Integer search and its local-rejection path are in 0.6.13-dev.

If a source-search expression fails its local typecheck or pure evaluation, the
history uses `gooo/joint-construction/v3` and retains a `rejection`. That
combination consumes one program attempt and receives no native caller score.
Local totals on that row cover only the already scored prefix. Search continues
within the original budgets, and saved replay rederives the rejected expression
and reason. [Runnable rejection example](../examples/caller-search-rejection/README.md).
Request cancellation, invalid source/selection, reconstruction failure, native
toolchain failure and compiled-program execution failure still stop the request.

## Reconsider complete multi-hole assignments

Gooo 0.6.14 development source also connects `source_fill` to caller-guided
construction. Its declared or grammar-derived assignments fill condition,
variable and return expressions together. The initial local winner is attempted
first, followed by source assignment order. Each completed combination is checked
and executed immediately within the existing whole-program budget.

The [budget helper example](../examples/caller-source-fill/README.md) includes
conditional assignment and a record result. Its v4 observation adds
`source_fill_index` selectors and `fill_candidates`. Every candidate records all
hole expressions, input/selected-source and plan digests, local training cases
and separate local holdout cases. Training and native caller cases guide selection;
holdouts never enter ranking or `COMPLETE_FINITE`. Those finite cases can all pass
while a holdout fails. Final evaluation remains a separate native execution.

`--fill-model` accepts the existing operation-classifier profile for initial local
fills; `--model` accepts the record-choice profile. Both are optional. Initial
model observations are retained; subsequent search and saved replay load no model.
Source-fill preflight still checks every assignment and stops on invalid ones.
The mixed v4 path preserves integer-search rejection rows and their scored-prefix
counts. Existing v1/v2/v3 observations keep their original meaning.

## Current boundaries

The graph supports 1..16 activities with 1..16 inputs and one result each.
An isolated activity needs no bind; every data-flow edge between multiple
activities remains explicit. The ordinary pure Integer/Boolean/Text body profile
is supported.
Source is bounded to128 KiB, generated function source to256 KiB, suites to32 KiB
and1..128 cases, and each decoded Text scalar to1024 UTF-8 bytes. Child stdout
and stderr retain the existing64 KiB limit. A total request has a60-second native
budget and each native run a2-second budget; cancellations and failures retain
their first stage. Optional learned source assembly currently accepts a single
Integer input and Integer result; multiple-input bodies connect these selected
results through checked ordinary source. The `run` command's registered operation
profile has its own input restrictions. The existing required-string field profile
now supports [record values and field traces](native-record-values.md). Feedback
across invocations, additional field types, calls and loops require further language work.

Related: [source assembly](source-assembly.md), [language direction](language-direction.ko.md),
[body generation](language/body-codegen.md).
