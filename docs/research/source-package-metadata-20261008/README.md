# Gooo source owns optional package metadata

Local observation: 2026-10-08, macOS arm64, Go 1.27.1. The candidate compiler is
clean source `b8b8ac3bd1899b3c3f709f838166e9d2157d37f2`; `build.json` retains its
identity. Its version string is still 0.6.10-dev, but this development source is
later than the published 0.6.10 binary. The frozen plan preceded model execution.

## Language and usage change

The workspace can omit `name` and `imports`. The compiler parses all listed Gooo
sources, checks their package names agree, and collects source imports in sorted
order without duplicates. Explicit assertions still use the existing consistency
checks. Package paths, source files and the entry remain in the manifest.

The public standard library at `27c0f6a0610f44fccd76758640933e1a8b0b99f8` supplied
the three library packages and the importing preview app. The prepared workspace
contains exactly those source bytes. Only package `name` and `imports` were removed
from the second manifest. The two complete normalized resolution results were
equal, while their manifest byte digests remained distinct.

## Native use with the existing own model

The model is the existing graph QAT bundle from Workbench
`19485bb64278ab8ac219e042f92af677ff3f5051`. Metadata SHA-256 is
`3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202` and weight SHA-256
is `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.

| Run | Selection cases | Attempts | Native expectations | New inference |
| --- | --- | --- | --- | --- |
| Explicit metadata, fixed ordering | 5/5 | 8 | 10/10 | 0 |
| Source-derived metadata, fixed ordering | 5/5 | 8 | 10/10 | 0 |
| Source-derived metadata, own graph model | 5/5 | 8 | 10/10 | 1 |
| Saved source-derived program | 5/5 retained | 8 retained | 10/10 | 0 |

These runs repeat one existing consumer on ten distinct evaluation input tuples.
The tuples were frozen previously and are disjoint from its five selection
examples. This is a compiler regression using the old program, not a new model
transfer experiment. The exact-number observer compared all delivered input and
output traces, expected values and generated-program hashes across the four runs.
They matched. No source choices, fixtures or weights were adjusted.

Input-only replay produced three actual records with `OBSERVED`, 0/0 correctness
expectations and zero new inference. Historical model calls remain in the saved
construction record; `result.replay.model_calls` and runtime counts describe this
invocation. Model prediction took 109,709 ns and preparation 0.796625 ms, with
2,096 bytes of resident tensors. These are one local observation; host CPU
increment and whole-process memory were not measured as a comparison.

## Tests and retained material

Before implementation, the new source-derived resolution test failed with
`PACKAGE_IDENTITY_UNKNOWN`. The targeted race run subsequently passed source
metadata, existing resolution and starter tests. They cover multiple source files,
explicit assertions, missing/cyclic dependencies, package/header errors, real
imported calls, relocation/replay and a Gooo policy resuming a partial helper.
Local static checks, the billing semantic example and exact-revision policy
checks also passed.

`observations.tar.gz` retains both workspaces, the original resolution/execution
records, input-only replay and the Go preparation/comparison programs. Extract
into a fresh directory. The comparison can be repeated with
`go run /path/to/extracted/verify.go /path/to/extracted`; place the retained
`build.json` in that directory first. No compiler executable or model weights are
included. `summary.json` retains construction details and actual input-only
outputs. `SHA256SUMS` covers the public observation files.
