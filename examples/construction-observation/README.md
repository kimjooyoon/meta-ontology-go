# Explain how a Gooo program was constructed

This example constructs an integer normalization followed by an increment.
`Normalize` searches for one expression; `Lift` scores three complete fills of
two holes. Their conditions, candidates and finite cases live in Gooo source.

From the repository root, using a compiler built from this revision:

```sh
gooo package execute --json \
  --cases examples/construction-observation/cases.json \
  examples/construction-observation/gooo.workspace.json > /tmp/construction.json

gooo package execute --json --construction-receipt /tmp/construction.json \
  examples/assembly-explainer/gooo.workspace.json
```

Pass `--go /path/to/go1.27.1` to each command if the Go executable on PATH differs
from the required toolchain. Neither command requires a model download.

The first command builds and runs the generated graph twice. Its two runtime
inputs specify four activity outputs: `-9 → 0 → 1` and `11 → 11 → 12`. The
construction suites and runtime expectations retain separate denominators.

The second command recomputes candidate scores from the saved source, then runs
the Gooo explanation tool twice. For a fill candidate scoring 0/3 when another
candidate already scored 3/3, Gooo returns `USE_OBSERVED_CANDIDATE`. IR search rows
show only the best result observed through that attempt. The observation's
`profile`, `view` and `input_index` explain which source stage and output a row
belongs to. Source fills run before composition, so fill rows appear first.

The explanation is an observation with 0/0 supplied runtime expectations. It
returns suggested operations as data. Selection cases are the scope of each
score; holdout observations remain outside these policy inputs. Replaying scores
establishes consistency with the saved source. Historical model timings and
the origin of the saved search order keep their original evidentiary limits.
