# `.gooo` formatter contract

This is a contract and evidence fixture for the formatter prototype. It is
deliberately independent of `internal/syntax`: a syntax adapter may change AST
shapes without changing the parser-neutral formatter boundary.

## Falsifiable hypothesis

**H-FMT-1 — representable semantic formatting is stable.** For a source file
whose AST can be adapted to the formatter document, formatting and parsing the
result again preserves the same semantic fingerprint:

```text
adapt(parse(format(adapt(parse(source))))) ≈ adapt(parse(source))
```

The hypothesis is falsified by any of these observations:

- an entity or activity stable ID changes;
- a `prov:used` or `prov:wasGeneratedBy` edge is added, removed, or retargeted;
- the formatted output cannot be parsed by the same surface grammar;
- an adapter or document error produces non-empty formatted source.

Declaration order, whitespace, and display names are presentation differences.
They must not affect the semantic comparison. The formatter must not infer
ordinary Go calls, authorization, or other domain relations.

## Minimal evidence fixture

[`positive.gooo`](../testdata/contract/formatter/positive.gooo)
contains two entities and one activity with intentionally irregular spacing.
The expected canonical output is
[`positive.golden.gooo`](../testdata/contract/formatter/positive.golden.gooo).
The expected semantic inventory is:

| Measure | Expected value |
| --- | ---: |
| semantic nodes | 3 |
| `prov:used` relations | 1 |
| `prov:wasGeneratedBy` relations | 1 |
| total semantic relations | 2 |
| unrelated semantic rewrites | 0 |
| output diagnostics | 0 |

The relation count is two: one activity input and one activity result. The
fixture is intentionally smaller than the billing example so a later AST, IR,
BX, or codegen test can report a precise delta.

## Counterexamples and negative cases

The contract manifest records expected outcomes without claiming that they have
already run on the integration baseline:

- [`positive.drifted.gooo`](../testdata/contract/formatter/positive.drifted.gooo)
  is parseable but changes an entity ID. Semantic equivalence must be false.
- [`negative-unknown-reference.gooo`](../testdata/contract/formatter/negative-unknown-reference.gooo)
  should produce `formatter.invalid-document` and no output.
- An adapter document with a custom activity ID should produce
  `formatter.unsupported-identity` because the initial surface grammar cannot
  write that identity back.
- A nil AST should produce `formatter.missing-ast`, no panic, and no output.

These cases distinguish “formatting failed safely” from “formatting silently
changed meaning.” Both are required evidence; neither is a success signal for
an unimplemented host stage.

## Pass, fail, and deferred criteria

An implementation **passes** this contract only when the positive fixture has
an identical semantic fingerprint after parse-format-parse and every negative
case has its named diagnostic or expected fingerprint mismatch with empty
output where specified.

It **fails** if it panics, emits source after an error, changes a stable ID,
changes the two-edge semantic inventory, or accepts an unknown reference as a
semantic declaration.

It is **deferred** when the syntax adapter, formatter entry point, or runnable
semantic fingerprint comparison is absent. Deferred evidence must use
`promotion_eligible: false`; it must not be reported as a formatter or
self-hosting pass.

The current integration baseline is in that deferred state. The contract is
ready for execution after the formatter prototype and syntax adapter are
available together.

## Reusable input/output contract

The adapter boundary is intentionally small:

```go
type ASTAdapter interface {
    Adapt(ast any) (*Document, Diagnostics)
}
```

`Document` contains `Package`, `Namespace`, and ordered declarations. An entity
requires a stable `ID`; an activity may omit `ID`, in which case its identity is
derived as `namespace://activity/kebab(name)`. Inputs and output resolve only to
declared entities. A custom activity ID is not approximated: it is rejected
with a diagnostic until the surface grammar can represent it.

The formatter result contains canonical `.gooo` source and ordered
diagnostics. If any adapter or validation diagnostic has error severity, source
must be empty. The semantic fingerprint contains stable IDs and the derived
`prov:used`/`prov:wasGeneratedBy` edges, while excluding spans, declaration
order, and display-only names.

The future syntax adapter should map `syntax.File` into this contract without
changing `internal/syntax`. The future semantic adapter should compare the
same inventory against `semantic.Graph.SemanticCanonical`; the formatter must
remain a presentation projection, not a second business SSOT.

## Host and evidence boundary

The Go-hosted formatter is the first executable host once its entry point is
available. A future gooo-hosted formatter may shadow it, but cannot promote its
own result. Both hosts must emit the same source, semantic, and diagnostic
digests for the pinned fixtures before comparison can become promotion
evidence. Until then, the manifest uses `evidence_status: deferred` and
`promotion_eligible: false`.

The machine-readable fixture index is
[`contract.json`](../testdata/contract/formatter/contract.json).
