# Read-only comparison of the original queue-batch observations

The three compressed inputs are byte-identical copies of the original public
queue-batch receipts from wiki commit `8d7ad7a31a1960909c0e0795886448a52ba42c27`.
They were produced once with public0.6.26/source8a759cba, using the released
choice model with SHA256
`aa81790884cd21ac9dd435fe5cf14b604cdd658f514f6460a526fc858ae07ab1`.
The original source and protocol were frozen before execution in wiki commit
`ce4413ddd01caf15b3dc8481f06cacd574adf8bc`. The public
[usage observation](https://github.com/kimjooyoon/meta-ontology-go/wiki/Small-Workflow-Pilot)
retains its six original commands, including failures of the first model choices.

This change adds a comparator and reads those existing observations. No model
fit, prediction, candidate search or native execution was repeated to produce
these comparisons. Earlier one-use studies remain intact.

| Comparison | Aligned groups | Changed outcomes | Changed expectations | Same expectation, regression |
| --- | ---: | ---: | ---: | ---: |
| batch16 → batch32 | 10 | 5 | 5 | 0 |
| batch16 → its saved replay | 10 | 0 | 0 | 0 |

`change.json` retains all exact before/after values; `change.txt` is the default
human-readable output. `saved-replay-delta.json` contains the second comparison.
The counts describe one synthetic workflow's ten input tuples, not customers,
financial savings, causal performance, or independent model evaluation.
No confidence score is used to label a regression.

## Recount

Build the development compiler as described in
[the comparison guide](../../workflow-outcome-delta.md). From this directory,
decompress each input to a fresh temporary folder and run:

```sh
shasum -a 256 -c SHA256SUMS
mkdir /tmp/gooo-outcomes-inputs
gzip -dc before.json.gz > /tmp/gooo-outcomes-inputs/before.json
gzip -dc after.json.gz > /tmp/gooo-outcomes-inputs/after.json
gzip -dc replay.json.gz > /tmp/gooo-outcomes-inputs/replay.json
/tmp/gooo-outcomes body-outcomes-delta --json \
  --before /tmp/gooo-outcomes-inputs/before.json \
  --after /tmp/gooo-outcomes-inputs/after.json
```

Replace the last input with `replay.json` to inspect the saved-replay comparison.
This reads existing evidence only. Structural declarations come from
`internal/outcomedelta/delta.gooo`; generated Go and JSON Schema are checked
against that source. Regression fixtures cover exact integers, reordered cases,
whole caller tuples, bound intermediate inputs, multiple ports, records,
conflicting duplicates, absent expectations, partial observations and faults.

## Repository integration

The first CI on source `f612c0ce` found the new declaration and two examples
outside the syntax inventory. Focused comparator tests had passed; the full
repository checks exposed this registration omission. The report declaration is
now a third source-bound receipt projection. Its Go and JSON outputs are checked
against the declaration, including missing or changed artifact fixtures. The
construction examples follow the existing `.gooo.fixture` convention and remain
byte-identical to the frozen source. Existing syntax cases and their obligations
remain in the registry. Original comparison records above are unchanged.
