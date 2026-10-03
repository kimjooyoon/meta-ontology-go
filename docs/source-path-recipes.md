# Small path recipes from Gooo source

`gooo/source-typed-path-recipe/v1` derives the typed base from a selected Gooo
`computes` body. A recipe names the few places where construction may choose
between two alternatives. This saves the caller from copying expressions,
statement indices and the activity ID into a second JSON representation.

Think of the body as a small assembly drawing. The recipe marks the joints that
can move; a local model can suggest their positions, and finite tests measure
the resulting behavior. Expansion uses the compiler and requires no training.

## Run the example

```sh
gooo body-codegen --json --activity Compose \
  --path-plan examples/body-codegen/path-recipe.json \
  examples/body-codegen/path-recipe.gooo.fixture
```

The source has three successive subtractions. The recipe exposes each operand
order, giving eight possible combinations and a budget of eight attempts.
Add `--path-model /path/to/model.json` to rank these choices with a compatible
local model. With no model, the existing declared-fallback ordering is
deterministic. This route expands into the existing full path document and uses
the same construction, feedback, finite completeness and native verification.

```json
{
  "schema": "gooo/source-typed-path-recipe/v1",
  "choices": [
    {
      "id": "first",
      "kind": "operand_order",
      "occurrence": 0,
      "intent": "첫 피연산자 순서를 뒤집는다. Reverse the first operands."
    }
  ],
  "test_cases": [{"input": 10, "expected": -3}],
  "max_attempts": 2
}
```

This shorter example exposes only the first subtraction. Its finite pass count
may be partial because the other two choices are absent. The checked-in recipe
exposes all three. Korean and English intentions remain explicit caller inputs.

## Five choices

`occurrence` is required and starts at zero. Each kind counts only its matching
sites in source order. A containing expression precedes a child that starts at
the same position. Unnecessary parentheses do not change this ordering.

| Kind | Selected site | Two alternatives |
| --- | --- | --- |
| `operand_order` | Binary expression | Original or reversed operands |
| `branch_layout` | `if` statement | Original or exchanged branch bodies |
| `local_reference` | Local variable read | Original name or `alternative_name` |
| `assignment_target` | Existing-local assignment | Original name or `alternative_name` |
| `root_order` | First of two adjacent top-level statements | Original or exchanged statement positions |

Name choices require a distinct alternative. Other choices reject
`alternative_name`. The typed plan checks the alternative's scope and type;
combination-specific failures remain recorded as type-rejected candidates.
Root order preserves every statement and only exchanges the selected pair.
Stable activity IDs continue to come from the source.

The source profile is one `Integer -> Integer` activity, with integer and boolean
literals, local declarations, assignment, explicit `if`/`else` blocks, returns,
and binary `+`, `-`, `*`, `<`, `<=`, `==`, `&&`, `||`. Negative integer literals
are supported. Calls, loops, other scalar types, unary variable negation and
other operators require future profile work. Explicit local Go types are limited
to `int64` and `bool`. The original projection and the expanded fallback must
both typecheck and describe the same typed body.

## Observation, export and execution

The recipe also works with:

- `body-context --plan recipe.json`: export the expanded source-bound model
  context without model calls.
- `body-codegen --path-observation`: add discriminating inputs from an explicit
  source oracle. The checked-in `path-recipe-observation.json` requests reuse and
  direct resolution if one candidate remains.
- `body-execute --path-plan recipe.json`: re-expand from the original source,
  replay the generation receipt, compile the emitted Go and observe supplied
  runtime cases.
- `body-path-stream`: place the recipe in `document`; workers expand it inside
  their existing bounded request lifecycle. Full documents and recipes can share
  a stream. Cancellation and output backpressure use the existing worker protocol.

Use the same source, recipe and activity throughout that chain. The receipt
hashes the **expanded document** and authoritative source; this version adds no
separate hash for original recipe bytes. A source edit can shift occurrence
indices, so recipes should be kept and reviewed with their source.

## Bounds and measurements

Source and recipe each have a 128 KiB limit. Recipes require 1–16 choices,
1–128 explicit finite cases and 1–64 search attempts. The request owns fixed
arrays of 128 expression slots and 128 statement slots, and lowering limits
semantic nesting to 16. Slices, compiler ASTs, emitted text and receipts still
allocate. Repeated `input` references share the required single input node.

With SDK v0.2.18, a body may leave its declared input unread. For example,
`return 2 - 3` can expose an operand-order choice and assemble `return (3 - 2)`
from finite expectations. The compiler appends the signature input only when
absent; it still consumes one of the 128 expression slots. Existing bodies keep
their expression indices. Try the checked-in constant example:

```sh
gooo body-codegen --json --activity Constant \
  --path-plan examples/body-codegen/constant-recipe.json \
  examples/body-codegen/constant-recipe.gooo.fixture
```

This uses ordinary deterministic search without training or a model. The SDK
continues to reject other unused expressions and missing or duplicate input
declarations. A literal-only body needs no recipe choices and can use ordinary
`body-codegen`; recipes continue to require at least one structural choice.

The source binding first uses the existing canonical body comparison. For
supported typed bodies that differ only in presentation, the receipt may use
`canonical_typed_body_tree/v1`: it keeps operators, child edges, local names,
scope and statement order while removing parentheses and normalizing integer
literal spelling. It does not equate different algebraic arrangements.

Recipe decoding performs additional validation and projection work before the
existing generation timer. Measure end-to-end process time when comparing a
recipe with a full document. Smaller authored JSON alone does not demonstrate a
latency, memory or model-accuracy improvement. Finite functional completeness
counts passed declared cases; general behavior needs broader independent cases.

Related: [incremental sessions](typed-path-sessions.md),
[useful observations](path-observation-loop.md), and
[language direction](language-direction.ko.md).
