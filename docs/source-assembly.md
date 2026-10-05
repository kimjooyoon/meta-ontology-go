# Author body assembly in Gooo

For record output bodies, [field-value assembly](record-field-assembly.md) adds
typed JSON cases and alternatives inside constructor fields. The Integer path
profile below keeps its existing `case` and operand/reference choices.

An activity can keep its baseline body, assembly intent, permitted paths and finite
expectations in one declaration. `assembling` is a typed syntax node, preserved by
bidirectional lowering into semantic IR. It expands through the existing recipe
arena and optional-model/TDD pipeline.

For multi-hole IR body generation, Gooo can own a finite set of complete
assignments or derive expressions and their combinations from a closed grammar.
A model ranks the complete assignments as one decision; it cannot add expressions
outside that set. Every selected hole is checked against the declared integer cases,
then the generated Go is typechecked. Without a model, the same bounded selector
uses deterministic scoring. The emitted Gooo source contains the chosen body and
does not retain a pending `source_fill` instruction.

```gooo
activity Lift(Integer) -> Integer computes `let base = __GOOO_BODY_HOLE_seed__
let increment = __GOOO_BODY_HOLE_step__
return base + increment` assembling {
    source_fill intent "Represent input plus one as a base and increment." {
        hole "seed"
        hole "step"
        candidate "add_one" {
            fill "seed" "input + 0"
            fill "step" "1"
        }
        candidate "double" {
            fill "seed" "input * 2"
            fill "step" "0"
        }
    }
    case "0" -> "1"
    case "2" -> "3"
}
```

To have Gooo derive the expressions and candidate assignments, replace the manual
`candidate` blocks with a bounded `derive` declaration:

```gooo
source_fill intent "Represent input plus one as a base and increment." {
    hole "seed"
    hole "step"
    derive grammar "integer-offset-constant/v1" max_expressions "8" max_candidates "16"
}
```

The closed grammar uses only the source's declared cases to enumerate integer
expressions. Gooo builds complete assignments in deterministic lexicographic order,
up to the assignment cap. Its receipt reports both expression grammar coverage and
the fraction of the complete assignment space actually enumerated. If the cap
truncates that space, the completeness dimension stays `PROGRESS`; a model can only
rank the enumerated prefix. See
`examples/body-codegen/source-ir-fill-derived.gooo.fixture` for a runnable example.

Run it with `gooo body-codegen --json --activity Lift
examples/body-codegen/source-ir-fill.gooo.fixture`. Add `--tiny-model
<model.json>` to route a manually declared finite decision through the local
compact Gooo model. The derived integer-expression grammar currently uses Laya
when configured or deterministic selection otherwise; the compact model's
operation-only output cannot represent its larger assignment vocabulary.
The receipt reports the selected candidate, per-case results, and completeness
dimensions; the percentage describes only these declared examples, not all inputs.

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
    [search hole <quoted-hole-id> grammar <quoted-grammar>
        intent <quoted-intent> max_candidates <quoted-count>]
    [source_fill intent <quoted-intent> {
        hole <quoted-id> ...
        candidate <quoted-id> { fill <quoted-hole-id> <quoted-expression> ... } ...
        | derive grammar <quoted-grammar> max_expressions <quoted-count>
            max_candidates <quoted-count>
    }]
    case <quoted-int64-input> -> <quoted-int64-expected>
    [holdout_case <quoted-int64-input> -> <quoted-int64-expected> ...]
    [attempts <quoted-budget>]
    [seed <quoted-seed>]
}
```

Canonical formatting puts each choice on one line. Whitespace/comments separate
fields. Quoted decimal numbers follow the existing policy grammar. Input and
expected values span int64. Formatting preserves decoded Korean/English intent.
`source_fill` requires 2–8 holes, and either 2–16 candidates with exactly one fill
per hole in declaration order, or one bounded `derive` clause with 2–16 expressions
per hole and 2–16 complete assignments. Manual candidates and `derive` are mutually
exclusive. It is mutually exclusive with path
choices, search, checkpoints, sampling seeds, and attempt budgets.

When holes need different expression types, declare a grammar for each hole:

```gooo
derive assignments max_candidates "16" {
    hole "condition" grammar "integer-predicate/v1" max_expressions "8"
    hole "result" grammar "integer-offset-constant/v1" max_expressions "2"
}
```

`integer-predicate/v1` derives `true`, `false`, and bounded comparisons between
`input` and the distinct training inputs. `integer-offset-constant/v1` derives the
bounded integer expressions described above. These are finite source-derived
grammars, not unrestricted Go or proofs over every integer. The receipt reports
retained expressions per hole, total grammar coverage, assignment count and any
assignments omitted by the cap. The compiler typechecks each complete assignment,
scores it against the declared cases, and can send the complete choices to Laya
for ranking; deterministic fallback uses the same scored choices.

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

## Declare a typed IR search in the Gooo source

The generated-expression search can also live in an activity's `assembling`
block, so the source owns its intent, candidate grammar, training cases,
withheld holdout cases and attempt budget. This mode does not require a JSON
sidecar or a hand-authored candidate list:

```gooo
activity ClampNegativeToZero(Integer) -> Integer computes "let output = input; if input < 0 { output = __GOOO_BODY_HOLE_floor__ } else { output = input }; return output" assembling {
    search hole "floor" grammar "integer-offset-constant/v1" intent "Return zero for negative integers and preserve zero or positive integers." max_candidates "16"
    case "-2" -> "0"
    case "-1" -> "0"
    case "0" -> "0"
    case "1" -> "1"
    case "2" -> "2"
    holdout_case "-9223372036854775808" -> "0"
    holdout_case "9223372036854775807" -> "9223372036854775807"
    attempts "8"
}
```

Run the source directly; `GOOO_LAYA_URL` optionally configures the sequential
chooser. If it is absent, the generated candidates are tried in their stable
Gooo order:

```sh
gooo body-codegen --json --activity ClampNegativeToZero \
  examples/body-codegen/ir-search-source.gooo.fixture
```

The `search` declaration supports the bounded
`integer-offset-constant/v1` grammar. `case` rows train candidate selection;
`holdout_case` rows are withheld from chooser requests and measured after
selection. Both remain part of the source's semantic identity. The receipt's
grammar coverage is scoped to expressions generated by that named grammar; it
does not claim intent understanding or correctness across all integer inputs.
`attempts` is an upper bound and the engine stops when it reaches either that
budget or the end of the generated candidate set.
This source-owned mode cannot be combined with path-choice fields, a second
external plan, or record-field assembly in the same activity.

To run the whole source-owned search path in one command, including a native
build and two executions of the selected program, use `body-search-run`:

```sh
gooo body-search-run \
  --source examples/body-codegen/ir-search-source.gooo.fixture \
  --activity ClampNegativeToZero \
  --cases examples/body-codegen/ir-search-runtime-cases.json
```

When `GOOO_LAYA_URL` is set, Gooo asks Laya to rank candidates in order and
validates each proposed choice against the training cases. Without Laya, it
follows the same stable candidate order. The JSON response keeps generation and
execution receipts together: `training_passed/training_total` measures declared
selection examples, `holdout_passed/holdout_total` measures the withheld source
examples, and `execution.observation.cases` measures the independent runtime
suite. The three denominators stay separate; passing finite examples does not
establish intent understanding or all-input correctness.

Related: [recipes](source-path-recipes.md), [language direction](language-direction.ko.md),
[worker](native-body-worker.md).

[Native body composition](native-body-composition.md) connects these selected
bodies to other checked scalar activities through source-declared binds. The
whole graph is compiled once and executed immediately with ordered intermediate
input/output traces and separate finite runtime expectations.

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
