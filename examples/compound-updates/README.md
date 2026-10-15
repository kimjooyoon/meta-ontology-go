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

Supported operators are `+=`, `-=`, `*=`, `/=` on initialized direct integer locals and integer fields of local records. String fields support `+=`. Division truncates toward zero. Update a local copy of an input; input parameters remain readonly. The native fault-origin extension below covers field division. Dynamic zero division on a direct scalar local still needs a separate observation repair.

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
