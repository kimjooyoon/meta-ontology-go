# Reuse a compiled Gooo graph for current inputs

`body-compose` can retain a compiled graph while executing an ordered series of
input/expectation suites. Gooo source still determines the functions and their
connections. The model is consulted during construction. Each runtime suite
executes the resulting program twice and records its current values.

It works like keeping an assembled tool on the workbench and feeding it the next
piece of material. The first execution builds the tool. Later executions use the
same tool and measure what happens to their own inputs.

## Try changing inputs and expectations

```sh
go run ./cmd/gooo body-compose \
  --source examples/body-codegen/record-field-assembly.gooo.fixture \
  --case-series examples/body-codegen/record-field-case-series.json \
  --out /tmp/gooo-field-series
```

Add `--model /path/to/model.json` to order the field choices with a compatible
local model. Omit it for deterministic construction. The three-suite example
contains different record inputs. Its expected named-output scores are 4/4, 2/2
and 1/2. The last suite deliberately gives a different expected text. The actual
record and label are retained, so that difference can feed the next comparison.

The response's `runtime_history` is in suite order. `runtime` is the most recent
execution. The directory contains `runtime-history.json`, `case-series.json`
and the ordinary source/composition/program files. `cases.json` contains the
first suite used during construction. Each runtime has its own suite fingerprint
and actual input/output trace.

## Compare first build and reuse

`--repeat 2` repeats the ordered suites in the same executor. The request supports
at most 16 executions in total. Each suite keeps the existing 1..128-case/32KiB
limit; a case-series document contains 1..16 suites and at most 512KiB.

For a single existing case file, use:

```sh
go run ./cmd/gooo body-compose \
  --source examples/body-codegen/record-field-assembly.gooo.fixture \
  --cases examples/body-codegen/record-field-assembly-cases.json \
  --repeat 3
```

The first history entry reports a current `build` and
`artifact.reused=false`. Subsequent entries report `artifact.reused=true`, the
original `artifact.source_build`, and their current run observations. A retained
native Go toolchain has its original version observation under
`toolchain_reference`. Original build/version costs remain separate from current
work. All runtime calls make zero model predictions; model calls belong to the
single construction recorded in `composition`.

Use `--composition /path/to/composition.json` with the series for a saved
construction. It makes zero new predictions and performs its first native build
in the new request, followed by reuse within that request.

## Lifecycle and Go callers

An `internal/bodyexecution.Executor` now also exposes `ExecuteComposition`.
It retains at most one scalar or graph executable and one Go toolchain
observation. The key contains generated Go, the graph driver, compiler/toolchain
identity, target platform and child environment. A changed program replaces the
previous workspace. Changed inputs and expectations execute with their own
finite scores on the retained program.

Calls use the existing cancellable executor gate. `Close` cancels active work,
waits for child completion and releases the workspace and scratch buffer. The
CLI closes the executor before returning. This retention scope is one owned
executor; a new command begins with its own first build.

Owned graph observations use `gooo/body-composition-runtime/v2`. Ordinary
single-suite execution retains the original v1 shape. The runtime history keeps
partial expectation matches as well as complete matches.
