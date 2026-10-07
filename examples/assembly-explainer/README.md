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
| `budget` | Declared number of candidate evaluations allowed |

Invalid count ranges return `FAIL_CLOSED`. An absent suite and an unscored set
remain `UNKNOWN`. An observed 0/N result with candidates still available returns
`PROGRESS`. The tool can point to a better recorded candidate, continue within
the budget, or request an expanded choice set. It returns the next operation as
data; consuming that operation is a separate application step.

These inputs are supplied observations. This small program does not read a
generation receipt or establish the observations' provenance by itself. Use the
original source and receipt when adapting a real assembly result. Changing the
source conditions changes the tool; no Go classification branch needs editing.

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
