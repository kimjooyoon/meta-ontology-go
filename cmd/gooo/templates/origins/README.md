# {{module}}: a Gooo origin reader

Start by running the four practice cases:

```sh
gooo package execute --json --cases cases.json gooo.workspace.json
gooo package execute --json --inputs inputs.json gooo.workspace.json
```

`inspect.gooo` owns the classification. It receives the same 25 primitive fields
as `field-input-locations` and returns the stable activity/record/field IDs,
original file coordinates, expression and unresolved reason. Change the Gooo
conditions and output fields to build your own report.

The bundled cases use fictional `demo.gooo` coordinates and `SAMPLE_SOURCE_ID`.
They exercise a declared alternative, a field target, a missing template and an
invalid range. They are practice inputs, with four explicit expected outputs.
`inputs.json` is an editable input with no expected-output accuracy score.

## Read an actual saved execution

Use a JSON execution receipt produced by a compatible Gooo development compiler:

```sh
gooo package execute --json \
  --origin-receipt /path/to/saved-execution.json \
  --origin-kind field-input-locations --origin-limit 9 \
  gooo.workspace.json
```

The consumer reads recorded rows without running their original producer or its
decision model. Inspect the origin-input report for total/returned rows, page
truncation and observation gaps. Use `--origin-offset 9` for the next page.

The result is in `result.runtime.traces[].deliveries[0].actual`:

| status | Meaning |
| --- | --- |
| `DECLARED_ALTERNATIVE` | A mapped source-owned record-field alternative |
| `FIELD_TARGET` | A mapped authored field name |
| `ORIGINAL_EXPRESSION` | Another mapped expression |
| `UNRESOLVED` | A retained location gap and its reason |
| `INVALID_LOCATION` | A claimed mapped row missing a usable location |

To start from a field rather than a row number, add `--origin-field-id` with its
exact semantic ID. `--origin-activity-id` names the activity containing the event;
`--origin-root-activity-id` names its entry root. The predicates are combined.
The report retains `source_rows` before selection, `total_rows` after selection
and the returned page size. Keep the same IDs when following `next_offset`.
Producer event/write numbers and full lineage case journals stay unchanged.

Ranges are UTF-8 byte offsets. A mapping describes a recorded location;
complete data/control ancestry and producer correctness need their own checks.
The starter creates no model, endpoint or network connection. The generated
`AGENTS.md` describes source editing and the installed compiler's inspection
commands. Keep declaration IDs stable when renaming your own display names.

Source and workspace identities use `{{module}}`. Move this project as one
directory; the generated workspace uses the local `inspect.gooo` path.
