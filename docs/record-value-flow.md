# Follow the value used by a record expression

Gooo records copy their values at a statement. A later write changes the working
record while an earlier saved record keeps its fields. Saving a scalar field
also keeps its value at that point. The optional source flow export makes these
relationships visible before model prediction or finite case execution.

```gooo
let current = input0
let saved = current
let previousState = current.state
current.state = "ready"
current.reason = saved.title + ":" + previousState + ":" + current.state
return current
```

This resembles working on a document beside an earlier photograph. Reading the
document and reading the photograph can produce different values even when the
field names match.

## Inspect the source graph

From a checkout with Go 1.27.2:

```sh
go run ./cmd/gooo body-context --value-flow --activity Select \
  examples/body-codegen/record-field-updates.gooo.fixture
```

The optional `value_flow` has schema `gooo/record-value-flow/v1`. Definitions,
copies, writes, reads, expression operations and both permitted choice
expressions are nodes in source order. Parents always precede a node. Record
fields retain stable field IDs. Local numbers follow declarations and lexical
scope; renaming a local preserves those relations.

Each node has a byte span within `computes`, or the saved planning `baseline`.
Alternative spans use `alternative:<choice ID>` and offsets within that
expression. Input-port nodes have their own view. The body digest describes the
normalized planning body, where `let` has the same-width `var` spelling; the
outer export also identifies the original Gooo source.

When both branches continue, a `join` keeps their possible definitions and the
condition that selects them. A branch that returns contributes no later value.
Execution guards also follow the continuation after an early return. Nested
returns can leave a union of live paths, recorded by `guard_join`. Source
constant conditions carry `known_truth`; supplied input values stay outside
this symbolic analysis. Record equality is represented by field comparisons.

The graph supports the current pure typed record body profile. It has at most
512 nodes, 64 simultaneous bindings, 16 nested conditions and six choices.
Record values use sixteen fixed field slots. Exceeding a bound, encountering an
unsupported relation, or stopping before a declared choice returns
`UNRESOLVED` with a reason and no partial graph. The usual context export keeps
its existing JSON shape unless this option or the new model contract is selected.

## Follow a pure Gooo helper

The graph follows the fixed, same-source helper activities already supported by
body generation. Each call connects its actual arguments to fresh
`call_parameter` definitions, analyzes the helper's body, and links its returned
value back through a `call` node. Record arguments and returned fields retain
copy-by-value relationships. Calls that occur only in an alternative expression
are included in the source closure too.

The optional `helpers` list binds each helper's stable activity ID, name and
original program digest. Helper body spans carry `activity_id` and offsets in
that helper's same-width-normalized `computes` body. Call nodes retain
`callee_id`; their spans describe the caller's expression, including an
`alternative:<choice ID>` when applicable. A local rename preserves the value
relationships, while source digests and affected byte spans still describe the
actual source.

Several return statements are joined using their execution guards. A return
inside a source-constant unreachable branch does not contribute to the helper's
result. Conditionals and nested calls retain their own scopes. No input values,
expected outputs or model predictions are needed to build this graph.

The 512-node bound applies to the expanded graph, and the 64-binding bound
includes live caller and callee frames together. The pure source closure keeps
its 32-activity and 16-call-level limits; nested conditional analysis retains its
16-level limit across calls. An exceeded bound returns the complete unresolved
status without nodes, choices or helper metadata from a partial expansion.

For example, `Classify` in the [filename example](../examples/text-operations/README.md)
uses `HasSuffix`, `HasPrefix`, `StripSuffix` and `ByteLength`. Its graph can follow
both return paths of `StripSuffix`, including its nested `HasSuffix` call:

```sh
go run ./cmd/gooo body-context --value-flow --activity Classify \
  --feature-version triple_record_field_flow_v2_shared_v1 \
  examples/text-operations/source.gooo.fixture
```

## Feed the relation to the small model

```sh
go run ./cmd/gooo body-context --value-flow --activity Select \
  --feature-version triple_record_field_flow_v2_shared_v1 \
  examples/body-codegen/record-field-updates.gooo.fixture
```

The new explicit SDK v0.2.24 contract prepares two fixed 16-slot ancestor-count
arrays per choice. Slots distinguish input fields, scalar inputs, copies,
field/scalar writes, previous choices, conditional joins, execution guards and
reads. Counts use stable field identity relative to the target field. Reachable
ancestors, including guard and join conditions, are counted once. Field names,
ordered expressions and complete Korean/English intent remain in the input.

Each field gets 64 expression slots, 32 origin slots and 160 intent slots.
Channels normalize independently; the three fields occupy 768 float slots.
The input bound is 1,024 UTF-8 bytes per part and 4,096 complete bytes. The full
graph digest remains in the source context receipt, separate from model features.
Input values and finite expected outputs are excluded from both graph and model
features. An optional expanded plan carries the finite cases separately.

`body-codegen --path-model` and `body-compose --model` dispatch a model with that
explicit contract to one origin-aware prediction before finite selection.
The retained worker shares immutable weights and gives each request its own
workspace. Saved selection replay reconstructs the source context with zero new
predictions. Unresolved flow or oversized context records the reason and follows
the existing deterministic candidate order.

The published v1 record model keeps its original contract and weights. This
implementation uses synthetic weights in ABI regression tests. New origin
weights need their own training and native-execution study. Current/saved
receiver pairs that shared the old expression array now produce different
origin arrays in the regression fixture. This measures representation, while
finite cases continue to measure functionality. Different graphs can share
ancestor counts, and the shared judge still scores the three fields independently.

Helper expansion extends the source relations available to the existing origin
projection. Its feature vocabulary is unchanged: call wrappers use the existing
read/other category, and helper IDs remain in the source graph. In particular,
the structural projection still merges several operators such as `&&` and `||`.
A resolved graph is evidence that the source relation was represented; its
compressed model input and prediction quality need their own measurements and a
separately versioned improvement. This change adds no trained weights.

[Sequential updates](record-field-updates.md)
· [Field assembly and finite completeness](record-field-assembly.md)
· [Ordered operators and source values in model input](record-graph-model-input.md)
· [Preregistered representation study plan](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/main/docs/local-value-origin-plan.ko.md)
