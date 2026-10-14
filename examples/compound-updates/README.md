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

Supported operators are `+=`, `-=`, `*=`, `/=` on initialized direct integer locals. Division uses integer truncation toward zero. Readonly input updates, unbound locals, record-field compound updates and other compound operators remain outside this source profile. A zero denominator stays unresolved or fails the deterministic/native evaluation obligation.

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
