# Source-owned runtime package surface

Three consumers in `gooo-go-ports` currently repeat the same eight expected Go
import paths. This Gooo program owns that expectation in `inspect.gooo` and
compares it with an independently collected observation. It uses existing record,
sequence, loop and helper-call support through the normal package runtime.

The eight paths belong to the recorded catalog comparison application. They are
Go packages, not eight versioned module dependencies. To adapt this example,
review the declaration and independently specify the application's expected
results. Deriving the declaration from today's `go list` output would make an
unexpected dependency look permitted.

## Execute and replay

From the compiler repository root, using an installed `gooo` binary:

```sh
gooo package execute --json --receipt-compression gzip \
  --cases examples/runtime-package-surface/cases.json \
  examples/runtime-package-surface/gooo.workspace.json > /tmp/runtime-surface.json.gz

gooo package replay --json --receipt-compression gzip \
  --receipt /tmp/runtime-surface.json.gz \
  --cases examples/runtime-package-surface/cases.json \
  examples/runtime-package-surface/gooo.workspace.json > /tmp/runtime-surface-replay.json.gz
```

For an observation, use `--inputs observation.json` instead of `--cases` in both
commands. The input has this existing package-runtime shape:

```json
{
  "schema": "gooo/body-composition-inputs/v1",
  "inputs": [{
    "example/runtime-package-surface:Inspect": {
      "origin_ref": "sha256:<digest of retained original command output>",
      "import_paths": ["an exactly observed import path"]
    }
  }]
}
```

The host collects non-standard packages with
`go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./cmd/compare`
inside the application module. Retain command, source revision, working directory,
Go version, build context and raw output before extracting nonempty lines. Preserve
their spelling and order. A cross-platform or tagged build needs its own observation.
`origin_ref` is an opaque caller-supplied reference; this program checks presence
and carries it forward. The adapter must bind it to the actual retained bytes.

The assessment is at `result.runtime.traces[].deliveries[].actual`. It includes
the declared paths, unchanged observed paths, counts and origin. An input-only
receipt has decision `OBSERVED`, finite denominator zero, and no model calls.
`OBSERVED_EXACT` inside the assessment describes matching package membership in
this build context. It grants no dependency-update or merge authority.

## Independent cases and their limits

The fourteen frozen cases check whole assessment records: the exact set, reversed
order, a missing package, an unknown package, repetition, equal-size substitution,
case differences, empty input, missing origin, repeated unknown paths, trailing
space, a partial set, a missing path hidden by repetition, and empty input without
origin. The fixture is deliberately separate from the Gooo declaration.

`missing_count` counts absent declared paths; `undeclared_count` counts observed
occurrences outside the declaration. Repetitions remain visible even when a
missing or unknown package determines status. Missing origin has precedence;
counts and original paths remain available. An empty or repeated declaration
produces `INVALID_DECLARED_SURFACE`.

Two source-mutation tests substitute an unknown path or repeat an existing path
without rewriting the fixture. Each must fail all fourteen expectations. The CLI
then returns a structurally valid `PROGRESS` receipt with exit code zero. Automation
must inspect decision and finite outcomes instead of using exit code alone.

This example adds no production Go comparison algorithm and no compiler command.
The existing runtime bounds collection size, text and transport independently;
it is intended here for a small application package surface. Source/generator
lineage and distribution-source rosters have separate owners and follow-up work.
The three downstream lists are still pending consumer migration.

The [development retrospective](../../docs/research/development-retrospective-20261025.ko.md)
records how this relates to source-owned changes and scoped provenance.
