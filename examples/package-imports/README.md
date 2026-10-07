# Construct imported Gooo activities and explain their results

The `core` package increments an integer. The importing `app` package returns
that result. Their Gooo bodies contain explicit holes; `body-plans.json`
declares the permitted expressions and expected construction cases for each
activity. The same compiler checks both plans and the package binding.

Run from the compiler repository root with a newly built `gooo`:

```sh
gooo package execute --json \
  --body-plans examples/package-imports/body-plans.json \
  --cases examples/package-imports/cases.json \
  examples/package-imports/gooo.workspace.json > execution.json

gooo package execute --json --construction-receipt execution.json \
  examples/assembly-explainer/gooo.workspace.json
```

Use `--go /path/to/go1.27.1` if the default Go executable differs from the
required toolchain. Both commands work with deterministic construction.

The first command retains each external plan together with its candidates,
selected body, source identity and native execution. The second reconstructs
those plans and passes whole construction-case counts to the Gooo explanation
tool. Its conditions and messages are defined in the tool's Gooo source.

| Activity | Candidate | Matched construction cases | Gooo suggestion |
| --- | --- | ---: | --- |
| `core.Normalize` | `increment` | 3/3 | Observe new inputs |
| `core.Normalize` | `identity` | 0/3 | Use the observed better candidate |
| `app.Main` | `identity` | 3/3 | Observe new inputs |
| `app.Main` | `zero` | 1/3 | Use the observed better candidate |

All four rows carry `profile: external_fill` and `view: scored_set`. Each plan
scores its two candidates before selection, so the best observed score is
already available to both rows. Their capacity is two candidates per plan.
Holdout scores and historical native successes do not enter these policy inputs.

The runtime case specifies two activity outputs for one root input, `7 → 8 → 8`.
That input was already used during construction. Keep its runtime result separate
from the two three-case construction suites and the four explanation rows.
The explanation is input-only `OBSERVED`, with no expected-output correctness
count. Its next-operation suggestions are returned as data.

## Reuse the same program

The saved plans also support fresh execution:

```sh
gooo package replay --json --receipt execution.json \
  --cases examples/package-imports/cases.json \
  examples/package-imports/gooo.workspace.json
```

Supply a new case or input file to use the program on different inputs. Replay
and interpretation make zero model calls. Older external-fill receipts missing
their plans need one construction run with the original `--body-plans` file.

## Recorded local use

The [published observation](../../docs/research/external-construction-20261008/summary.json)
uses clean compiler `8a48369e`. Construction matched the two named activity
outputs in 1.03 seconds; Gooo interpreted the four candidate rows in 0.40 seconds.
Each command performed two native runs and zero model calls. The root input
overlaps construction, so fresh-input correctness remains unmeasured here.
This is one sequential local run with uncontrolled warm caches. Raw receipts,
plan identities, outputs and reproduction commands accompany the summary.
