# Rejected expression continuation — 2026-10-09

The clean compiler `de5aba5b5103dd6e2f1c73afc9c3e09c7b46ba08` stopped on a
constant-zero denominator before reaching the next valid expression. Its original
output and failure are retained in `observations.tar.gz`. The scalar source,
construction and evaluation inputs were frozen in `c0bddd56` before the fix.
The mixed fixture was committed with the implementation before its model run.

The clean candidate is `ae3890c241575d6bbbe2a9d730fd015d36b97ccd`, Go 1.27.1,
decision runtime v0.2.26-experimental. It reports the existing development version
0.6.12; these changes are not in the public 0.6.12 release archive.

| Program / order / budget | Combination attempts | Local rejections | Native program attempts | Evaluation |
| --- | ---: | ---: | ---: | ---: |
| Scalar / fixed / 5 | 3 | 1 | 2 | 4/4 |
| Same scalar / fixed / 2 | 2 | 1 | 1 | 1/4 |
| Mixed / fixed / 40 | 24 | 8 | 16 | 4/4 |
| Same mixed / own model / 40 | 17 | 8 | 9 | 4/4 |

Each native attempt uses the existing two-execution comparison. A locally
rejected expression consumes the combination budget without a native caller
score. The eight mixed rejections are different combinations containing the same
zero expression. The two mixed completed routes selected byte-identical Gooo.
Saved scalar and model histories replayed their original failures and finite
outputs with zero new inference. The partial scalar outcome remains 1/4.

These are two programs, four construction runs and two saved replays. The scalar
evaluation roots differ from the construction root; zero is also its local helper
example. The mixed evaluation roots differ from its two construction roots.
Those facts do not establish independence from model-training data.

The unchanged workbench model is
`models/graph-chooser-20261008/all-data-demonstration/qat_ternary/model.json`:
metadata SHA-256 `3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`,
weights SHA-256 `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.
It orders record choices once during initial construction; the scalar grammar
order remains fixed. No weights were trained or changed. Host CPU utilization
and comparative latency were not measured.

## Recount and validation

Extract the archive to a fresh directory, copy the adjacent `build.json` into
that directory, and run:

```sh
go run ./docs/research/caller-search-rejection-20261009/observe.go \
  /tmp/rejection-observations examples/caller-search-rejection
```

The Go recount uses exact JSON numbers and independently stored input/expectation
files, including `9007199254740993`. It verifies caller counts, unscored
rejections, final source equality and saved actual values. `summary.json` and
`verification.txt` are its output. Compiler tests also rederive original local
values, rejected-candidate identity, bounds and saved history.

Focused old/new joint tests passed. `go vet ./...` and serial package race checks
for bodycodegen, bodyexecution and the CLI passed. The scored-prefix regression
was added and checked separately afterward; final-commit CI remains separate.
The native variable-zero failure test confirms that native execution errors still
stop the request. This work does not claim continuation after native panics or
support for all-invalid initial local construction.
