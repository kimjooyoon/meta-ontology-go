# Two packages with their own Result records

Both packages declare `Result`. The read package's record carries `value` and
`note`; the write package's record carries `scaled`. Their stable entity and
field IDs describe two different types.

The program increments an input, extracts the resulting integer, and doubles it
into the write package's record:

```sh
gooo package execute --json \
  --cases examples/package-record-namespaces/cases.json \
  examples/package-record-namespaces/gooo.workspace.json
```

Use Go 1.27.2, or pass `--go /path/to/go1.27.2` for the native build. Four input
rows check twelve activity outputs, including an integer above 2^53. The command
builds the selected graph and runs it twice.

Each package resolves names within its own declarations and direct imports. The
compiler gives colliding record names distinct execution names derived from
stable IDs. `result.program.entity_aliases` retains each original package, name,
ID and lowered name. Source files, string values, local variables and field names
stay attached to their original meaning. Explicit bindings still require matching
stable entity IDs across their producer and consumer ports.

The same record ID can be declared under different names when its ordered field
IDs, names, types, presence and cardinality agree. Those aliases share one lowered
record definition. Conflicting shapes for one stable ID produce a diagnostic.

The body profile supports pure scalar and record computations with the current
field types. Record constructor scoping also covers source-owned body-fill
expressions, field alternatives, saved planning baselines and external body-fill
candidates. Only activities in the entry's producer chain have their bodies
prepared for execution. A model can rank supported declared choices after this
deterministic name resolution; this example uses no inference.

The [recorded run](../../docs/research/record-namespaces-20261007/summary.json)
links the compiler revision, source digests and raw receipt: twelve named outputs
matched across four input rows, with two fresh native runs and zero model calls.
This measures the example's finite behavior and record identity separation.
