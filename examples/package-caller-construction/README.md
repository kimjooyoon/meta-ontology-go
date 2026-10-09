# Caller guided package construction

A library function can pass its own examples and still give the caller the wrong
result. Here a retry tool imports `PlanBudget` from `tools/budget`. All viable
candidates pass its local example at `used=0`. At the caller's boundary
`used=limit=8`, their behavior differs.

`gooo package construct` uses that caller result to reconsider the imported
function. It retains package identities, original examples, rejected candidates,
native faults and the selected program. The published 0.6.17 binary supports
construction with evaluation cases. Build this development revision to use the
new input-only path shown below.

## Run the example

From the repository root, with Go 1.27.2:

```sh
go build -o /tmp/gooo-package ./cmd/gooo
/tmp/gooo-package package execute --json \
  --cases examples/package-caller-construction/construction-cases.json \
  examples/package-caller-construction/gooo.workspace.json > baseline.json

/tmp/gooo-package package construct --json \
  --construction-cases examples/package-caller-construction/construction-cases.json \
  --cases examples/package-caller-construction/evaluation-cases.json \
  --attempts 6 examples/package-caller-construction/gooo.workspace.json > construction.json

/tmp/gooo-package package construct --json --receipt construction.json \
  --cases examples/package-caller-construction/evaluation-cases.json \
  examples/package-caller-construction/gooo.workspace.json > replay.json
```

Use new output filenames to keep earlier observations. The source files stay
unchanged. Omitting `--json` prints a short summary.

The baseline keeps `late_unbounded` and matches **0/1 caller expectations**.
Construction with five attempts stops with **1/4 evaluation expectations**;
six attempts reach **4/4**. Its history includes two local rejections, a native
zero-divisor fault and the successful `bounded` assignment. Each completed
candidate is executed immediately, twice, before the next is considered.
The evaluation includes the exact integer `9007199254740993`.

The construction example is consumed feedback. The four evaluation inputs are
different caller inputs and run after selection. A source-local holdout is
recorded separately. These are finite observations; the input-separation count
describes caller inputs and does not establish unseen model-training data or
unseen values at every nested call.

## Combine two libraries with the compact model

`mixed.workspace.json` imports the budget function and a second library with
three record choices. The declared combination space is `6 × 8 = 48`.

```sh
/tmp/gooo-package package construct --json \
  --construction-cases examples/package-caller-construction/mixed-construction-cases.json \
  --cases examples/package-caller-construction/mixed-evaluation-cases.json \
  --attempts 48 --model /path/to/graph-chooser/model.json \
  examples/package-caller-construction/mixed.workspace.json > mixed.json
```

The existing [workbench model](https://github.com/kimjooyoon/gooo-ecosystem-workbench/tree/main/models/graph-chooser-20261008)
orders record choices; `--fill-model` accepts a compatible operation classifier
for initial source-fill selection. Later attempts follow the source-bounded
combination order. Without model options the order is deterministic. Provider
environment variables are not used by this command.

An initial local observation with the unchanged 2,096-byte tensor model took
41 attempts, compared with 48 without it; both reached 4/4. This is the existing
all-data demonstration model and a related example. Package lowering changes the
combination enumeration order, so this result does not inherit the six-attempt
result from the earlier single-source example.

## Saved construction and limits

Replay reconstructs the current package image and its original caller examples,
then re-executes **all saved attempts** and the selected program on the supplied
evaluation cases. It makes zero new model calls and needs no model files.
`replayed_from_sha256` binds the input receipt. Prior evaluation measurements
remain in that receipt; this invocation reports fresh measurements.

A relocated copy of the same workspace is supported. Changed package sources,
imports, stable identities, entry or caller cases require new construction.
`--receipt` excludes new construction examples, budgets and model options.
`COMPLETE_FINITE`, `PARTIAL_FINITE` and `OBSERVED` receipts can be replayed. Process or
compiler failures return an error and retain available construction history.

The command supports the existing typed package closure with source record
choices, integer IR search and source fills, within 1–64 program attempts.
External fill plans and custom assembly-policy workspaces remain on the
`package execute` / `package resume` routes. A saved package construction is
consumed by `package construct --receipt`; it has its own receipt schema.

## Run actual inputs

Once a program has been selected, an application usually has inputs without
known answers. `inputs.json` supplies four such rows, including an integer above
JavaScript's exact-number boundary. Run the saved construction on them:

```sh
/tmp/gooo-package package construct --receipt construction.json \
  --inputs examples/package-caller-construction/inputs.json \
  examples/package-caller-construction/gooo.workspace.json
```

The six-attempt program prints `3`, `-10`, `-9007199254740994`, and `1`, one JSON
value per line. To save the full observation, add `--json > observation.json`.
That receipt can be reused with either `--inputs` or `--cases`. Fresh construction
also accepts `--inputs`, together with the original labelled construction cases
and an attempt budget. These actual inputs do not influence candidate selection.

Input-only receipts report `OBSERVED` and bind the file with `inputs_digest`.
The current input observations contain no expected values or pass flags, and
their finite evaluation counts are zero.
`result.construction.decision` still says whether the consumed construction
examples reached `COMPLETE_FINITE` or remained `PARTIAL_FINITE`. Both kinds can
produce actual values; observation alone does not establish their correctness.

JSON retains native faults and blocked dependencies as observations. Plain mode
stops with a nonzero exit when the entry has no value and points to `--json` for
details. Values already printed belong to preceding successful input rows.
