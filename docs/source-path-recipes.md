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

## Common comparisons and negated conditions

The bounded typed-body profile also accepts Go-style `>`, `>=`, `!=` and boolean
`!` in source recipes. It lowers these spellings into existing typed operations:

```text
a > b   → b < a
a >= b  → b <= a
a != b  → (a == b) == false
!test   → test == false
```

The source-binding receipt uses `normalized_condition_typed_body_tree/v1` when
this normalization participates in a comparison. It still typechecks both
projections and compares all local references, statement order, branch bodies,
and remaining operators. This admits the four condition spellings without
expanding the runtime's operation set.

The [condition assembly example](../examples/body-codegen/condition-path-assembly.gooo.fixture)
combines comparison, inequality and negation with two local values, an assignment
inside an `if`, an alternate branch, a return expression and three declared
assembly choices. Its Korean and English recipe and two separate finite suites
are checked in beside the source. Run it without `--model` for deterministic
construction; supply a compatible tiny three choice `--model` to measure its
selection path. The emitted program is independently compiled and checked
against its runtime cases by `body-path-run`.

The source profile is one `Integer -> Integer` activity, with integer and boolean
literals, local declarations, assignment, `if`/`else` blocks and `else if` chains, returns,
and binary `+`, `-`, `*`, `<`, `<=`, `==`, `&&`, `||`. Negative integer literals
are supported. Calls, loops, other scalar types, unary variable negation and
other operators require future profile work. Explicit local Go types are limited
to `int64` and `bool`. The original projection and the expanded fallback must
both typecheck and describe the same typed body.

### Conditions in sequence

An `else if` chain uses the existing nested typed branches. For example, this
body limits a value to 0 through 10:

```gooo
activity Clamp(Integer) -> Integer computes "if input < 0 { return 0 } else if input < 10 { return input } else { return 10 }"
```

The checked-in `condition-chain.json` exposes the two branch layouts and the
upper comparison's operand order. It works with deterministic search or the
existing three-choice model. `branch_layout` occurrence 0 selects the outer
condition; occurrence 1 selects the next condition. The source order is retained
when the compiler lowers `else if` to `else { if ... }`. Existing variable scope,
the 16-level nesting bound and the shared 128-slot limits still apply.

```sh
gooo body-codegen --json --activity Clamp \
  --path-plan examples/body-codegen/condition-chain.json \
  examples/body-codegen/condition-chain.gooo.fixture
```

The emitted Go uses explicit nested blocks. Typed-tree source binding accepts
this presentation change while keeping condition order, local names, values and
branch bodies. Scope and return checks apply before any model call.

## Observation, export and execution

The recipe also works with:

- `body-context --plan recipe.json`: export the expanded source-bound model
  context without model calls.
- Add `--include-plan` to explicitly include `expanded_plan`, the typed source
  body and its declared alternatives. Its JSON digest equals
  `context.original_plan_sha256`. This lets experiments inspect operation order
  through the compiler's existing lowering. The export includes caller-authored
  intentions and names; choose public fixtures when publishing it. Test cases,
  expected results and feedback are excluded from `expanded_plan`. Source binding
  and plan validation finish before this optional field is emitted; failed or
  cancelled exports contain no plan. This flag performs no selection or training.
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

## Whole-candidate order judge (SDK v0.2.19)

The experimental [order judge](https://huggingface.co/asketeddy/gooo-order-judge-tiny-v1)
can be passed through the ordinary `--path-model model.json` flag. It sees the
complete original intent and the actual ordered operations in eight composed
candidates. Source expansion and binding precede loading; one local prediction
precedes finite candidate tests and selected Go emission. Its fixed neighboring
`weights.bin` file is 16 KiB and is checked against the closed model metadata.

The supported shape is `let v=input; v=op(...); v=op(...); return v`, using one
`root_order` and two `operand_order` choices. Operators are add/subtract/multiply;
leaves are local, input or integer constant; model constants are -16..16. Input
text is the declared choice intents joined with newlines, at most 512 UTF-8 bytes.
All declared choice orderings are retained. The search attempts up to the smaller
of the document budget and eight, reusing exact equal operation descriptors.

The result's `body_paths.whole_candidate_judgment` records all candidate ranks,
descriptors, prediction timing and equivalence skips. `search` retains evaluated
bodies and finite outcomes. The request may finish partially at a small budget;
`body-execute` reconstructs the selected source and compiles/runs its actual Go
projection, including cases that fail the supplied expectation.

This initial model route supports unseeded, uninterrupted bounded search.
Explicit sampling, batch and feedback options return a recorded error. An
unsupported body or feature bound also returns an error before prediction; input
is never silently shortened. Leave out `--path-model` for ordinary deterministic
search across the compiler's broader supported profile. The retained generator
loads the immutable model once and uses separate workspaces per request.

With SDK v0.2.20, the order route compiles its eight legal candidates once per
prepared plan and captures the model identity once. `NewTypedPathGenerator`
retains at most one such plan per generator. Every call still binds the source,
checks the full plan digest, makes a fresh prediction and evaluates its own cases.
Changing the plan replaces the retained entry; changing only cases or the attempt
budget reuses programs while producing fresh results. Concurrent misses may
prepare independently and never wait on another request's preparation.

`body_paths.whole_candidate_preparation` records the bound plan digest, whether
preparation was reused, the eight-candidate count and acquisition time. Tensor
bytes in model information exclude retained program arenas and source strings.
A fresh CLI process prepares once; reuse across calls requires a retained Go
generator. SDK search timing and complete compiler generation timing are measured
separately in the [public experiment](https://github.com/kimjooyoon/gooo-neural-decision-experiments/tree/main/publication/order-prepared-sdk-20261003).
