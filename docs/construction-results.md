# Reading a construction result

Development source adds `--format text` and `--format markdown` to `body-construct`.
The public 0.6.27 binary predates these options. JSON remains the default and keeps
the original shape. Every format uses the same construction and evaluation result;
formatting adds no model call, candidate evaluation, compilation or program run.

```sh
gooo body-construct --source examples/caller-guided-construction/main.gooo.fixture \
  --entry Main --construction-cases examples/caller-guided-construction/construction-cases.json \
  --cases examples/caller-guided-construction/evaluation-cases.json \
  --attempts 2 --out first-construction --format text
```

Choose a new output directory. Its JSON, selected Gooo and generated Go files are
complete regardless of the display format. Use `--format markdown` to copy the
table into a project discussion. Large JSON values in displayed failure details
are abbreviated at 120 characters; exact integers are kept as JSON numbers.

## The three populations

- **Selected local checks** come from the source bodies used in the selected
  combination. They do not include every possible candidate or input.
- **Caller expected outputs used for selection** were consumed while constructing
  the program. A later passing candidate does not erase the first candidate's score.
- **Evaluation expected outputs** belong to the separate suite run after selection.
  Each expected activity output counts once. One input row can check several
  activities, so this denominator can exceed the number of rows.

The report shows matched, mismatched, faulted, blocked and unobserved expected
outputs. An unobserved toolchain failure is reported as unmeasured rather than a
measured 0% score. An observed zero-match suite retains 0/N. Faults in activities
without an expected output are also reported, separately from the score.

Input overlap counts unique **caller root-input tuples**, plus duplicate rows.
An input absent from caller construction can still have appeared in a local source
example or model training. The report makes no held-out accuracy claim from this
overlap count. Candidate uniqueness and arbitrary-input correctness remain outside
this summary, even when every supplied check passes.

## Saved construction

```sh
gooo body-construct --source first-construction/original.gooo \
  --construction first-construction/construction.json \
  --cases examples/caller-guided-construction/evaluation-cases.json \
  --format markdown
```

Saved construction replay still rechecks the recorded history and runs the current
evaluation suite. It needs no model file. The report marks the historical
construction, whether replay succeeded, and any replay failure before evaluation.
It does not turn a failed command into a successful one. Use the JSON view for all
attempts, source identities and complete values; the short report shows at most
eight failed, faulted or blocked evaluation deliveries and counts the rest.
