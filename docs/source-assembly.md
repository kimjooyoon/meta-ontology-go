# Author body assembly in Gooo

An activity can keep its baseline body, assembly intent, permitted paths and finite
expectations in one declaration. `assembling` is a typed syntax node, preserved by
bidirectional lowering into semantic IR. It expands through the existing recipe
arena and optional-model/TDD pipeline.

```gooo
package offsets
namespace offsets
entity Integer id "offsets://integer"
activity Offset(Integer) -> Integer computes "return input - 2" assembling {
    choice "offset-order" operand_order at "0" intent "2에서 입력을 뺀다. Subtract input from two."
    case "-1" -> "3"
    case "4" -> "-2"
    attempts "2"
}
```

Initially `computes` is the baseline. The choice permits reversing one subtraction. The
cases cause the search to select `2-input`. A local model can rank the alternatives
before finite checks. Omitting the model gives deterministic bounded search.

## Commands

```sh
gooo body-codegen --json --activity Qualified \
  examples/body-codegen/source-assembly.gooo.fixture
gooo body-context --include-plan --activity Qualified \
  examples/body-codegen/source-assembly.gooo.fixture
gooo body-path-run --source examples/body-codegen/source-assembly.gooo.fixture \
  --activity Clamp --cases examples/body-codegen/source-assembly-clamp-cases.json \
  --repeat 2 --timing --out /tmp/gooo-source-assembly
```

Use a fresh output directory. The file runner saves source, generation, emitted Go,
independent cases and runtime/timing observations. It builds and immediately runs
each selected projection. A repeated identical projection can reuse the owned
native artifact; source replay and runtime observations remain fresh.
`body-path-run --verify-timing --out <directory>` checks saved timing and file bindings.

Add `--path-model /path/to/model.json` to `body-codegen`, or
`--model /path/to/model.json` to the file runner/worker. Existing model bounds apply.
The shared three-choice model supports both checked-in activities. The
whole-candidate order judge retains its narrower two-operation profile. This
authoring feature requires no new weights or training.

Separate execution accepts `body-execute --source <source> --generation
<generation.json> --cases <cases.json>`. The source supplies its plan. A worker
request can omit `document` for an activity with `assembling`; use the existing
`execution_cases` envelope with `--execute`. Parallel workers retain bounded queues
and request-owned arenas and workspaces.

## Grammar

```text
activity ... computes <quoted-or-raw-body> assembling {
    [baseline <quoted-original-body>
     picked <quoted-choice-id> -> <quoted-option-label> ...]
    choice <quoted-id> <kind> at <quoted-occurrence>
        [alternative <quoted-local-name>] intent <quoted-intent>
    case <quoted-int64-input> -> <quoted-int64-expected>
    attempts <quoted-budget>
    [seed <quoted-seed>]
}
```

Canonical formatting puts each choice on one line. Whitespace/comments separate
fields. Quoted decimal numbers follow the existing policy grammar. Input and
expected values span int64. Formatting preserves decoded Korean/English intent.

| Kind | Source site selected by `at` | Additional field |
| --- | --- | --- |
| `operand_order` | Binary expression | — |
| `local_reference` | Local read | Required `alternative` |
| `assignment_target` | Local assignment | Required `alternative` |
| `branch_layout` | Conditional | — |
| `root_order` | First of two adjacent root statements | — |

Occurrences start at zero in source order. Parent expressions precede children at
a shared position; `else if` keeps condition order. A saved checkpoint keeps these
coordinates relative to `baseline`, even when selected statements move. Scope/type
checks and checkpoint/body equivalence precede prediction.

The body profile supports one `Integer` input/result, local integer/Boolean values,
assignments, supported binary conditions and nested branches. Limits are 128 KiB
source, 128 expression slots, 128 statement slots and nesting depth 16. Declare
1–16 unique choices, 1–128 cases and one 1–64 attempt budget. Intent has 1–512 UTF-8
bytes per choice, IDs 1–128 bytes. An optional seed has at most 512 UTF-8 bytes and
requires a model for sampling. Model context budgets may be narrower.

## Meaning and measurement

Intent, alternatives, cases and budget participate in both bidirectional and core
semantic fingerprints. Editing them changes the contract and retains the activity's
stable ID. Formatting and detached cloning preserve it. Selections and runtime
observations remain in their existing result records. A checkpoint also carries
the original planning body and ordered picked labels through those same IR/BX paths.

An embedded assembly owns the complete plan. Omit external `--path-plan`, `--plan`
and stream `document`; combining them returns an error. Library generation, context
export and replay check the expanded document against the declaration. Activities
without the block retain external recipes/full documents and ordinary generation.

Expanded documents and original sources retain existing SHA256 bindings. Contract
verification reuses the checked baseline projection and prepares its typed plan.
Measure full request time when comparing authoring forms; no speed gain is assumed.

Finite selection and independently compiled execution are separate measures.
`Qualified` declares five selection cases. `Clamp`'s independent runtime suite
covers seven inputs, including int64 extremes and both range boundaries. Counts
describe supplied cases. Wider natural-language understanding, arbitrary bodies,
calls, loops and additional types remain development tasks.

Related: [recipes](source-path-recipes.md), [language direction](language-direction.ko.md),
[worker](native-body-worker.md).

## Continue developing from the selected body

`body-codegen --json` returns `source` (Go) and, for source assembly, `gooo_source`
(the complete reusable Gooo file). Its receipt records
`source_format: gooo/source-assembly-checkpoint/v1` and the selected source digest.
Only the chosen activity's `computes` token and `assembling` span change.

```sh
gooo body-codegen --json --activity Qualified \
  examples/body-codegen/source-assembly.gooo.fixture > /tmp/generation.json
gooo body-realize --source examples/body-codegen/source-assembly.gooo.fixture \
  --generation /tmp/generation.json --out /tmp/gooo-realized
gooo body-codegen --json --activity Qualified /tmp/gooo-realized/realized.gooo
```

The directory contains `original.gooo`, the exact `generation.json`, `realized.gooo`
and `realization.json`. The command reconstructs the declared selection and finite
observations without loading a model. It preserves a partial finite score as well
as a complete one. Older source-assembly generation records are replayed using
their original body-only format, then upgraded; both source digests are recorded.

```gooo
activity Offset(Integer) -> Integer computes "return (2 - input)" assembling {
    baseline "return input - 2"
    picked "offset-order" -> "layout_reverse"
    choice "offset-order" operand_order at "0" intent "2에서 입력을 뺀다. Subtract input from two."
    case "-1" -> "3"
    case "4" -> "-2"
    attempts "2"
}
```

The baseline is a coordinate map; `computes` is the implementation now in use.
Picked labels identify how that map produced the implementation. Name choices
retain both names when the selected read changes; interacting choices retain the
original palette even if one combination fails scope/type checks. Export and
typed generation verify that the working body matches the recorded selection.

Labels are `layout_forward/reverse` for operands and branches,
`reference_first/second`, `assign_first/second`, and `schedule_forward/reverse`.
`baseline` requires exactly one picked label per choice, in declaration order.
The complete output remains bounded to 128 KiB. A new baseline starts a new plan:
edit `computes`, remove the old `baseline`/`picked` fields, and align the choices
and cases with that body. Re-generating an existing checkpoint runs fresh bounded
search; known selections do not skip testing or guarantee broader intent coverage.
