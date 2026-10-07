# A language tool written in Gooo

This independent package explains a finite assembly observation and returns a
next operation. Its field types, conditions and Korean messages live in
[`main.gooo.fixture`](main.gooo.fixture). The compiler lowers and executes that
source through the existing native package runtime.

## Give the tool an input

From the compiler repository root:

```sh
gooo package execute \
  --inputs examples/assembly-explainer/inputs.json \
  examples/assembly-explainer/gooo.workspace.json
```

The command prints one JSON entry result per input row. The first row asks about
a proposal matching 1/3 cases when an observed candidate matches 3/3:

```json
{"state":"PROGRESS","next_operation":"USE_OBSERVED_CANDIDATE","message":"더 많은 사례를 만족한 후보가 기록되어 있습니다."}
```

`inputs.json` uses `gooo/body-composition-inputs/v1` and an `inputs` array of
named package-activity inputs. It contains no expected answer. Add `--json` to
retain the complete source, generated Go, input/output traces and two-run replay
receipt. The outer receipt says `OBSERVED`; its finite expectation counters are
0/0 and expectation evidence remains `UNKNOWN`. A `PASS` value returned by the
tool describes the supplied assembly counts, not independent verification of
this invocation or the underlying program's whole input domain.

## Input contract

| Field | Meaning |
| --- | --- |
| `matched` | Cases matched by the candidate being explained |
| `total` | Declared cases in that same suite |
| `best` | Largest matched count among the observed candidates |
| `scored` | Candidates actually scored against that suite |
| `budget` | Effective scoring capacity after the candidate cap and observed unscored attempts |

Invalid count ranges return `FAIL_CLOSED`. An absent suite and an unscored set
remain `UNKNOWN`. An observed 0/N result with candidates still available returns
`PROGRESS`. The tool can point to a better recorded candidate, continue within
the budget, or request an expanded choice set. It returns the next operation as
data; consuming that operation is a separate application step.

These fields can come from caller input or the construction receipt adapter
below. Changing the source conditions changes the tool; no Go classification
branch needs editing.

## Interpret an actual construction receipt

```sh
gooo package execute --json \
  --construction-receipt docs/research/domain-tools-20261007/documentation-model.json \
  examples/assembly-explainer/gooo.workspace.json
```

The adapter reconstructs saved source-owned record choices, IR searches and body
fills, plus body fills with retained external plans, including candidate scores. Each scored candidate becomes one input
row for the Gooo entry. Counts use whole construction cases; they do
not use the historical runtime counters or the number of matching record fields.
The observation carries the original receipt hash, activity, candidate mask,
attempt index and source-declared budget in `construction_input`.

The example records a model's first candidate matching 2/3 cases, then a candidate
matching 3/3. Gooo returns `CONTINUE_CANDIDATES` followed by `OBSERVE_NEW_INPUTS`.
For record choices and IR search, `view: attempt_prefix` means `best` includes
only candidates already observed at that point. `scored` counts completed
evaluations. The declared attempt budget is capped by the candidate space and
reduced by unscored attempts seen so far. Type or evaluation failures retain
their reason and `scoring_completed: false`; they receive no policy input and
their zeroed count fields carry no case measurement. A measured 0/N retains its
positive denominator. `input_index` maps scored observations to policy outputs.

For record choices, `proposed` identifies the mask in the retained model
prediction, while `selected` identifies the final candidate after finite
continuation. They can identify different rows. Deterministic enumeration has
no model-proposed row. Interpretation checks the retained ranking and candidate
scores without rerunning that historical model prediction.

For body fills, every candidate is evaluated before selection. These rows use
`view: scored_set`; all have the final scored-set size and observed best score.
The declared capacity is the retained plan's candidate set size. `source_fill`
identifies a source-owned contract; `external_fill` identifies the plan retained
in the saved body-fill step. Replay reconstructs and scores that exact plan.
`selected` and
`proposed` identify the final and proposed candidates. Holdout scores stay out of
both policy input forms. [Run both profiles](../construction-observation/README.md).

This interpretation makes zero model calls. The target Gooo tool is compiled and
executed twice with no supplied expected answers, so its outer result remains
`OBSERVED` with 0/0 runtime expectations. The returned recommendations describe
the recorded construction sequence; they do not execute a repair or establish
correctness on new program inputs.

The current adapter handles source-owned record choices, integer IR search and
source-owned or externally planned body fills, with at most 128 scored rows per request.
External fills require the saved `external_plan`; regenerate older receipts with
the original plan if that field is absent. Other assembly profiles return an
unsupported-profile error. [Run the imported-package example](../package-imports/README.md)
to pass two external plans through native execution and Gooo explanation.
It passes the five-field input contract to the target manifest's entry; this
example uses a single ordinary entry activity. `--construction-receipt`,
`--cases` and `--inputs` are mutually exclusive.

The [recorded interpretation](../../docs/research/domain-tools-20261007/observation-summary.json)
binds the original model-run receipt, compiler revision, derived inputs and actual
Gooo outputs. Its raw receipt retains the two fresh native runs. That historical
record uses the first adapter schema; current observations use
`gooo/construction-input/v2` with explicit profile and input mapping.

## Check the tool separately

```sh
gooo package execute --json \
  --cases examples/assembly-explainer/cases.json \
  examples/assembly-explainer/gooo.workspace.json
```

The ten named expectations cover unobserved and measured-zero cases, a better
candidate, remaining/exhausted budget, complete finite coverage, invalid counts,
and exact integers above 2^53. Runtime results apply to those ten cases.
The manifest selects one entry activity; no synthetic producer or binding is
needed. Multiple-input entries use explicit `.input0`, `.input1`, etc. keys.

If the installed Go differs from the compiler's required toolchain, pass the
matching executable with `--go`. The package runtime is bounded to pure typed
bodies, 1..128 input rows and two fresh native executions per request.
