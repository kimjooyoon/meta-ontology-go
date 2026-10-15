# Updating a local value

This program adjusts an amount, applies a branch-specific update and divides the result. Its arithmetic and branch remain in Gooo source.

```gooo
let amount = input0
amount += input1
if input2 {
 amount *= 2
} else {
 amount -= 3
}
amount /= 2
return amount
```

Each compound update has two value origins: the preceding local value and the right-hand expression. Native execution records the old-value read before the new write, with source coordinates for the authored left-hand variable. The path recipe expands the update into a binary assignment while retaining a normalization receipt.

Supported operators are `+=`, `-=`, `*=`, `/=` on initialized direct integer locals and integer fields of local records. String fields support `+=`. Division truncates toward zero. Update a local copy of an input; input parameters remain readonly. Native fault observation covers both local and field division.

Multi-input case keys name the ports individually:

```json
{
  "schema": "gooo/body-composition-cases/v1",
  "cases": [
    {
      "inputs": {"Adjust.input0": 8, "Adjust.input1": 6, "Adjust.input2": true},
      "expected": {"Adjust": 14}
    }
  ]
}
```

The four fresh finite inputs used for the published source are `(8,6,true)`, `(19,4,false)`, `(-13,-4,true)` and `(-4,7,false)`, with outputs `14`, `10`, `-17`, `0`. These cases establish observed behavior and provenance for that population. [Origins](../../docs/observations/compound-updates/README.md).

## Updating one field while retaining its origin

`Summarize` in `fields.gooo` uses named inputs and a local record copy:

```gooo
let summary = 기록
let delta = 기록.step + 1
if 더하기 {
 summary.count += delta
} else {
 summary.count -= delta
}
summary.count *= 2
summary.count /= 3
summary.title += "✓"
return summary
```

The updated count retains both its preceding count and the local `delta` path back to `기록.step`. The title retains its original title. `untouched` keeps its own origin. Input display names are separate from the canonical `input0` connection port.

```sh
gooo package execute --json --cases examples/compound-updates/fields.cases.json \
  examples/compound-updates/fields.workspace.json
gooo roundtrip --json examples/compound-updates/fields.gooo
gzip -dc examples/compound-updates/recorded-fields.json.gz > /tmp/gooo-record-fields.json
gooo package execute --json --origin-receipt /tmp/gooo-record-fields.json \
  --origin-shape field-local-columns --origin-offset 9 --origin-limit 4 \
  examples/local-field-bridge/trace.workspace.json
```

`roundtrip` checks source/semantic round-trip laws for the record program. The command with `--origin-receipt` reads a saved runtime without executing the producer again. The Gooo reader reports input roots for the four returned fields. The local/field bridge is needed when an operand passes through a scalar local such as `delta`.

The first four native cases matched their expected records and retained 16 compound writes. The first reader exposed a port-name error; the repaired producer and reader retained the same four cases. Separate new division cohorts contain two normal outcomes and three recorded faults, including a division inside the right-hand operand. These are finite observations, with zero model calls. [Saved observations](../../docs/observations/record-field-updates/observations.json).

The saved four-case runtime belongs to the archived first source, where the record example followed `Adjust` in one file. The public record example now has its own registered file and profile. A readonly comparison confirms identical generated Go for `Summarize`; the saved runtime keeps its original source hash. Registration v86 adds one case and one source path (124 cases / 142 paths), preserving the earlier registry entries and denominator files.

Field observation currently covers flat records with required integer, boolean and string fields. A dynamic zero divisor in this profile records `ZERO_DIVISOR`, the operation's source identity, and an unfinished write. An explicit constant zero denominator is rejected during generation. Indexed targets and other compound operators remain outside the update syntax.

## A divisor supplied at execution time

`Divide` in `main.gooo` uses `amount /= 분모`. Both operands are read before the
write can complete. A zero divisor retains `ZERO_DIVISOR`, `assignment: "/="`,
`variable: "amount"`, the original source expression and its byte range. The
interrupted local write remains incomplete. A division inside the RHS has its
own source site, so its failure is distinct from the enclosing update.

```sh
gooo package execute --json --cases examples/compound-updates/division.cases.json \
  examples/compound-updates/division.workspace.json
```

The two runnable cases produce `19` and `-19`. A separate recorded case with
`149 / 0` records a fault instead of completing the expected numeric output.
The new helper/branch and mixed local/field cohorts bring this observation to
12 distinct cases: 7 normal outcomes and 5 faults. The first revision failed
with an unobserved panic; its five-case repair retains the same denominator.
Model calls are zero. [Original attempts and results](../../docs/observations/local-division-origins/observations.json).

### Read an execution fault from the terminal

`division.inputs.json` supplies two ordinary inputs followed by a zero divisor:

```sh
gooo package execute --inputs examples/compound-updates/division.inputs.json \
  examples/compound-updates/division.workspace.json
```

This diagnostic example prints the first two values and exits with failure on
the third input. The error identifies the activity, operator, operands and Go
projection expression. A local `/=` with an exact retained source binding also
shows its Gooo right-hand expression, byte range and source hash. The selected
Gooo source and the Go projection have separate byte coordinates. `--json`
retains the complete execution record, including original-file annotations when
available.

The display checks read the saved execution above: one complete terminal view
and twelve source-binding variants, with zero producer executions. Missing,
ambiguous or mismatched source bindings leave the Gooo location undisplayed.
[Display observations](../../docs/observations/fault-diagnostics/observations.json)
retain the command results and sample terminal message.
