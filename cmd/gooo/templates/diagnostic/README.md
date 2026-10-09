# A diagnostic tool written in Gooo

This project turns observed counts and a diagnostic detail into a next-step
message. `diagnostics.gooo` declares the record fields, conditions, assignments,
three construction choices and five selection cases. `app.gooo` imports that
tool and formats its result. Package and semantic identities belong to
`{{module}}`.

## Build once, then reuse

From this directory, using Gooo and Go 1.27.2 for your host:

```sh
gooo package execute --json --cases cases.json gooo.workspace.json > execution.json
gooo package replay --receipt execution.json --inputs inputs.json gooo.workspace.json
```

The replay prints three values, including:

```text
"partial: missing branch result [repair-and-replay]"
```

Edit `inputs.json` to use your own counts and detail. The four Diagnose inputs
are matched outputs, total outputs, rejected candidates and explanatory text.
Input-only replay reports the values it observes. To check expected outputs,
use `--cases cases.json` in place of `--inputs inputs.json`. This supplied suite
has four input rows and eight activity-output expectations, separate from the
five source-declared construction cases.

If the default Go executable differs, add `--go /path/to/go1.27.2` to either
command. The complete directory can be moved together; replay checks the saved
choices against these source files and performs two fresh native runs. Changes
to source declarations require a new construction receipt.

## Optional small model

Deterministic ordering works immediately. A compatible local three-field model
can rank the same declared choices during the first construction:

```sh
gooo package execute --json --assembly-model /path/to/model.json \
  --cases cases.json gooo.workspace.json > execution.json
```

The [own shared QAT model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary)
is one compatible model. It has 2,072 parameters. Gooo evaluates proposed
constructions against the declared cases and records partial attempts. Replay
uses the saved program with zero new inference calls; prior model calls remain
historical observations in its receipt.

The diagnostic returns a suggested operation as data. Applications can consume
that suggestion; the tool itself only computes its output. The named cases
establish finite behavior, with broader task coverage still open.

Recipe origin: the MIT-licensed [Gooo ecosystem workbench](https://github.com/kimjooyoon/gooo-ecosystem-workbench/blob/d821c346/recipes/diagnostics.gooo).
