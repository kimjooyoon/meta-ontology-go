# Caller-guided multi-hole bodies

Gooo 0.6.14 development source adds whole-program selection for `source_fill`.
The [release guide](../../docs/releases/0.6.14-dev.md) explains the records and
reuse commands. Build this revision with Go 1.27.2:

```sh
go build -o /tmp/gooo-source-fill ./cmd/gooo
/tmp/gooo-source-fill body-construct \
  --source examples/caller-source-fill/budget.gooo.fixture --entry Main \
  --construction-cases examples/caller-source-fill/construction-cases.json \
  --cases examples/caller-source-fill/evaluation-cases.json \
  --attempts 3 --out /tmp/gooo-budget-construction
```

`PlanBudget` has a boundary condition and an assignment at the limit. Three
complete assignments satisfy its one local training example. The caller example
distinguishes late/unbounded, early/wrong-cap and bounded behavior. Two attempts
leave a partial result; three can satisfy the supplied local and caller cases.
The separate final suite includes an exact integer above 2^53.

The observation is `gooo/joint-construction/v4`. Each `fill_candidates` row
retains source/plan identities, every hole expression, original expectations and
actual values. `local_passed/local_total` includes training cases only.
`holdout_cases_passed/total` and the holdout rows remain separate, even if
holdouts fail. `COMPLETE_FINITE` describes the supplied training and caller cases.

Replay reconstructs the candidate history and executes final inputs with no
new inference:

```sh
/tmp/gooo-source-fill body-construct \
  --source /tmp/gooo-budget-construction/original.gooo \
  --construction /tmp/gooo-budget-construction/construction.json \
  --cases examples/caller-source-fill/evaluation-cases.json
```

The mixed fixture adds a three-choice arithmetic helper to this budget program,
giving 3 × 8 = 24 combinations. Its expected outputs were frozen before the
own-model observation. This is a composition experiment, not another independent
application. The existing graph QAT model may order the record choices using
`--model`; fill assignments follow their local winner and then source order.
An optional compatible operation-classifier model uses `--fill-model` for
initial fill preparation. The graph model is a different profile. Neither
model runs during final execution or saved replay.

The current source-fill preflight still checks all declared assignments and
stops on invalid ones. Integer `search` slots retain their separately recorded
local rejection behavior. Neither mode creates arbitrary statements: the source
owns the body structure, typed holes and finite assignment space.
