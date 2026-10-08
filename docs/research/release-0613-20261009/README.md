# Gooo 0.6.13 candidate regression

Clean candidate source: `141b38d64c41f2a799db2c708652de2adc4864ac`, with the
implementation and frozen [plan](PLAN.md) committed before building. It reports
0.6.13-dev, Go 1.27.1 and decision runtime v0.2.26-experimental. These observations
precede final PR CI, merge, four-platform readiness and public release.

## Native platform check

The local macOS arm64 witness completed PASS/EXACT. Two builds and archives
matched. Existing division, retry, filename, package, source-graph and caller
construction checks passed, followed by the new integer-search regression.

The new regression tried three combinations: caller mismatch, locally rejected
zero divisor, and a matching `-input` expression. Only two programs executed
natively. The original local zero expectation remained intact; four final inputs
and saved replay matched, including input `9007199254740993`. Replay retained
the same selected program and made zero new predictions. The structural release
case denominator remains separate from these additional language observations.

The first local invocation supplied an unsupported target label and stopped at
`TOOLCHAIN_RELEASE_TARGET_UNKNOWN` before execution. `platform-invalid-target.log`
retains it. The corrected invocation uses the tool's configured `macos-15` target
label; this is a local host observation. Actual CI platform runs remain separate.

## Own model reused with the candidate

The unchanged graph QAT model and frozen mixed program from the feature
observation were used again:

| Program / order / budget | Combination attempts | Local rejections | Native program attempts | Evaluation |
| --- | ---: | ---: | ---: | ---: |
| Scalar / fixed / 5 | 3 | 1 | 2 | 4/4 |
| Same scalar / fixed / 2 | 2 | 1 | 1 | 1/4 |
| Mixed / fixed / 40 | 24 | 8 | 16 | 4/4 |
| Same mixed / own model / 40 | 17 | 8 | 9 | 4/4 |

Fixed/model completed source is byte-identical. Both saved replays match with
zero new inference. The Go recount verifies every original caller/evaluation
value exactly and compares all six outputs, their selected sources, attempted
selectors and rejections against the earlier clean ae3890c2 run. It also compares
the platform witness's two raw search outputs with the direct scalar command.

These repeat the same two programs with unchanged weights and expectations.
The scalar evaluation includes its helper's local zero example. Evaluation-root
separation does not establish model-training independence. CPU utilization and
controlled latency comparisons were not measured.

## Reproduce the recount

Extract this `observations.tar.gz` and the prior
[feature archive](../caller-search-rejection-20261009/observations.tar.gz) into
separate fresh directories, then run:

```sh
go run ./docs/research/release-0613-20261009/observe.go \
  /tmp/release-0613-observations examples/caller-search-rejection \
  /tmp/prior-rejection-observations
```

The current `build.json` is adjacent to the archive; copy it to the extracted
candidate directory. `summary.json` and `verification.txt` are the recount output.
Full toolchainrelease and toolchaincli package tests passed, as did focused
version/release tests and `go vet ./...` before the observation utility was added.
The final source is checked separately in CI. No compiler binaries or model
weight copies are committed in this observation directory.
