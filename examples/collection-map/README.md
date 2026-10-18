# Transform a bounded collection in Gooo

This experimental body profile applies a source-declared activity to each item:

```gooo
return Decisions{values: map_values(input.values, IsPackageLabel)}
```

The example's Gooo activity defines the package-label rule: 1–64 ASCII bytes,
starting with a lowercase letter, followed by lowercase letters, digits or `-`.
Its result stays in input order. Duplicate labels retain separate positions.

From the repository root, with Go 1.27.2:

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/collection-map/cases.json \
  examples/collection-map/gooo.workspace.json > /tmp/gooo-map-execution.json
```

The two example cases expect five Boolean decisions and an empty list.
Execution includes native replay and uses no model. The generated program
evaluates the input expression once and reserves one 256-slot scalar array. The returned slice has length and capacity
equal to the input length, so unused slots cannot be read by reslicing. Native
origin observation uses one array sized to the actual input length.

## Source and origin

The runtime receipt retains the map expression, the mapper's stable activity ID,
the actual input length, and each attempted element's read interval. Reading a
mapped result links to its construction and element index. Mapping a slice keeps
the original input indices; identical values do not collapse identities.

```sh
go run ./cmd/gooo package execute --json \
  --origin-receipt /tmp/gooo-map-execution.json --origin-kind elements \
  examples/native-origin-reader/elements/elements.workspace.json
```

An element's source expression names the mapper. Its read interval describes
what was read while evaluating that element. A mapper fault preserves the
attempted prefix and blocks dependent execution through the existing fault path.

## Bounds

- Input and output are scalar collections: Integer, Text or Boolean, up to 256 items.
- The mapper is a fixed unary `computes` activity, resolved locally or through a package import.
- Static accounting reserves `256 × (1 + mapper call bound)` calls per map occurrence,
  including the surrounding loop multiplier. The existing 4,096-call limit applies.
- Observation retains the existing 4,096-construction and 65,536-element budgets.
  A full observation buffer reports partial origin coverage; value computation continues.
- Text field byte limits and per-case execution limits continue to apply.

Record-valued items, captured anonymous functions, filtering and parallel
evaluation are future work. This fixture exercises the experimental execution
entry point. The general syntax-readiness denominator is unchanged by this example.
