# Select source-owned collection values

The Gooo body declares both the input collection and the rule that keeps an
item:

```gooo
activity KeepNonempty(Text) -> Boolean id "urn:collection-filter:keep-nonempty" computes `return len(input) > 0`
activity Select(Labels) -> Labels id "urn:collection-filter:select" computes `return Labels{values: filter_values(input.values, KeepNonempty)}`
```

Run the three named cases from the repository root with Go 1.27.3:

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/collection-filter/cases.json \
  examples/collection-filter/gooo.workspace.json > /tmp/gooo-filter-execution.json
```

The expected outputs cover duplicates and Korean text, all rejected values,
and an empty input. The predicate runs once per input item in source order.
The result keeps its original order and duplicate positions. No model is
needed; a saved receipt can be replayed without calling one.

For these three named cases, the current implementation reports 3/3 expected
outputs, 7/7 input reads with resolved origins, three recorded filter
constructions, and zero model calls. These counts describe this fixture only.

The runtime records a `filter_value` read for every input item and a
`filter_values` construction element for each retained output item. Each
output element has an `input_index` within the collection passed to the
filter. Its source read retains the original input collection index, including
when the collection is sliced. Rejected inputs still have read observations;
they have no output construction element. The generated loop's own `if` is
not reported as a source-authored condition.

The current profile accepts Integer, Text, and Boolean collections of up to
256 values and a fixed unary source activity returning Boolean. Nested records
and captured anonymous predicates remain outside this profile. For text,
`len` measures UTF-8 bytes as in Go.
