# Let a caller reconsider a locally complete branch

Development source after the public 0.6.23 release adds typed-path bodies to
`body-construct`. Build that source with Go 1.27.2; the existing 0.6.23 release
continues to support its recorded record/search/fill construction contracts.

`Choose` has two permitted branch layouts. Both return zero for its local
`0 -> 0` example, so that example cannot distinguish them. `Main` supplies the
construction expectation `3 -> 6`. Gooo tries the locally selected body first,
executes the caller, and then tries the other branch within the source budget.
The separate evaluation file contains positive, negative and exact int64 inputs.

```sh
gooo-dev body-construct --source examples/caller-typed-paths/main.gooo.fixture \
  --entry Main --construction-cases examples/caller-typed-paths/construction-cases.json \
  --cases examples/caller-typed-paths/evaluation-cases.json --attempts 2 \
  --out /tmp/gooo-caller-paths

gooo-dev body-construct --source /tmp/gooo-caller-paths/original.gooo \
  --construction /tmp/gooo-caller-paths/construction.json \
  --cases examples/caller-typed-paths/evaluation-cases.json
```

Use a new output directory. The JSON receipt keeps local cases, consumed caller
feedback, final evaluation inputs and new inference counts in separate fields.
Its `gooo/joint-construction/v7` schema adds `typed_path_mask` and
`path_candidates`, including the source palette, actual local results and
rejected combinations. Earlier v1..v6 receipts keep their existing semantics.

## Model use and deterministic continuation

`model.gooo.fixture` adds three choices: branch layout, comparison operands and
subtraction operands. It accepts a compatible own three-choice model via
`--model /path/to/model.json`. Initial generation records its model calls and
chosen body. Subsequent joint candidates make no model calls: they start with
that selected mask, then use increasing distance from the source fallback,
breaking ties by numeric mask. Saved replay also makes zero new calls.

Every helper's source `attempts` includes the initial selected mask. The whole
program's `--attempts` separately bounds combinations across helpers. A rejected
combined declaration order consumes an attempt and receives no caller score.
Caller success cannot erase a failing local case. Different evaluation inputs
measure finite behavior for those inputs; this does not establish independence
from model training or accuracy across the whole language.

`mixed-model.gooo.fixture` is for the existing own record model. Its three
record-field choices receive inference; the one-decision typed helper records
the representation decline and proceeds deterministically. Use its
`mixed-construction-cases.json`, `mixed-evaluation-cases.json`, and `--attempts 16`.
The receipt makes that split explicit, rather than reporting model ranking for
every body. No new model or larger training set is required for this example.

The current typed recipe spells arithmetic negation as `0 - input`. Ordinary
Gooo supports `-input`; that spelling has not yet been added to recipe expansion.
