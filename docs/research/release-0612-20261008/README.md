# Gooo 0.6.12 candidate regression

The implementation and frozen [plan](PLAN.md) are clean source
`2246ec35710486c4baede7653a24519053917d7b`. The candidate reports 0.6.12-dev,
Go 1.27.1 and decision runtime v0.2.26-experimental. These are local candidate
observations; public release, installation and four-platform CI are separate.

## Native platform and caller behavior

The local macOS arm64 platform witness completed PASS/EXACT. Its two builds and
two archives matched. Existing division, retry, filename, package and source-graph
checks passed alongside the added caller construction.

The new command tried two whole-program combinations. Both kept the helper's
original zero example at 1/1; actual caller results changed from 3 to 6, with
scores 0/1 and 1/1. The selected program then matched all four evaluation rows,
including `9007199254740993 -> 18014398509481986`. Saved history replay reproduced
those four values and the same generated program with zero new predictions.
One evaluation root was consumed during construction and three were different.
The platform artifacts retain fresh and saved command outputs.

## Existing own model used with the candidate

We reused the public three-choice source, cases and independently trained graph
QAT model from the feature observation. The four runs repeat one known program.

| Order and budget | Whole-program attempts | Evaluation | New inference |
| --- | ---: | ---: | ---: |
| Fixed, 1 | 1 | 1/7 | 0 |
| Own model, 1 | 1 | 7/7 | 1 |
| Fixed, 8 | 8 | 7/7 | 0 |
| Own model, 8 | 1 | 7/7 | 1 |
| Saved model history replay | prior attempt retained | 7/7 | 0 |

Completed fixed/model routes chose byte-identical Gooo source and returned the
same actual values. Every attempted helper still matched its one local example.
The evaluation includes one consumed caller input; zero also occurs in the
helper's example. Those counts do not establish unseen-task accuracy.

The two predictions took 41.750 and 42.291 µs. Model setup took 0.240791 and
0.172792 ms, with 2,096 resident tensor bytes. These are individual readings;
host CPU utilization and incremental process memory were not measured. No
training or weight change occurred. Replay ran after deleting the temporary
model copy; the original workbench weights remain available.

## Retained files

- `build.json`: exact candidate identity.
- `summary.json` and `observe.go`: exact-number recount of local and caller
  values, actual fixed/model source equality and saved evaluation equality.
- `observations.tar.gz`: native platform records, independent example inputs,
  all four model/fixed command outputs, selected source directories and replay.
- `release-tests.log`: release witness, CLI conformance and full CLI race tests.
- `verification.txt`, `platform.log` and `checksums.sha256`: local check results
  and retained-file digests.

The compressed set excludes compiler binaries, platform binary archives and
model weights. Extract it to a new directory and run `go run observe.go <directory>`
from this research directory to recount the saved readings. This reads historical
observations; the platform witness and saved CLI replay perform native execution.
