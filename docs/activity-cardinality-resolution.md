# Activity cardinality resolution

`gooo graph resolve-activity` is a compiler-native graph operation selected by
two independent public consumers. The selection evidence is pinned to
[`gooo-link v0.2.0-dev`](https://github.com/kimjooyoon/gooo-link/releases/tag/v0.2.0-dev).
Selection is not implementation evidence, so the conformance denominator keeps
the earlier `NOT_IMPLEMENTED` observation and tests the new implementation
separately.

## Command

```sh
gooo graph resolve-activity model.gooo --namespace billing --name PayOrder
```

The selector accepts `--namespace`, `--name`, and `--id-prefix`. Supplied fields
are combined with logical AND, and at least one field is required. Only semantic
IR nodes whose kind is `Activity` can match. Results are sorted by stable ID.

## States

| Occurrences | Decision | Reason | Next operation |
| ---: | --- | --- | --- |
| 0 | `UNKNOWN` | `ACTIVITY_NOT_FOUND` | `DECLARE_OR_WIDEN_ACTIVITY_SELECTOR` |
| 1 | `CLOSED` | `ACTIVITY_UNIQUELY_RESOLVED` | `USE_RESOLVED_ACTIVITY` |
| 2+ | `REFUTED` | `AMBIGUOUS_ACTIVITY_BINDING` | `NARROW_ACTIVITY_SELECTOR` |

Zero matches are explicitly `DIRECT_MISSING`; they are not converted to a
fixed point. More than one match is a contradiction to unique binding, not an
unknown value. Every claim retains stage, step, reason, next operation, and one
Munchausen proof category.

The command exits successfully only for one match. All three states emit the
same JSON schema, so CI can inspect counterevidence without parsing stderr.

## Quantified conformance

The meta model declares 12 activities and the denominator declares 12 matching
cells: 4 `FOUNDATION`, 4 `COHERENCE`, and 4 `REGRESSION`. Indicator classes are
3 `OUTCOME`, 5 `DRIVER`, and 4 `GUARDRAIL`. CI resolves all 12 activities one at
a time, then observes zero and many selectors, deterministic replay, graph hash
preservation, peak RSS, elapsed time, and repository writes.

These are observations, not claims of language completeness or utility. The
primitive becomes useful only when released consumers replace their duplicate
shell cardinality logic and preserve their domain-specific differences.
