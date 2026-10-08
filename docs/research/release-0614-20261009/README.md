# 0.6.14 native candidate observation

The frozen plan preceded the release changes. This local observation uses the
actual archive produced by native readiness at clean commit
`ab55beb3063cbc879a9acaa9405b3e6efc001e51`, Go 1.27.1, version 0.6.14-dev.
The archive reproduced byte for byte across two builds. Its SHA-256 is
`bce97cd0f17e2646fee8fa398c4b74d10964dd952084a2a807c008c4b29e27b7`.
The receipt's `macos-15` label names the configured target; execution was on the
local macOS arm64 host. Public release status belongs to the GitHub release page.

| Original program / mode | Program attempts | Final original cases | New model calls |
| --- | ---: | ---: | ---: |
| Budget, two attempts | 2 | 1/4 | 0 |
| Budget, three attempts | 3 | 4/4 | 0 |
| Budget, saved replay | 3 | 4/4 | 0 |
| Mixed, fixed order | 24 | 4/4 | 0 |
| Mixed, same own model | 3 | 4/4 | 1 |
| Mixed, saved replay | 3 | 4/4 | 0 |

These are two related frozen programs, four fresh constructions and two saved
replays. The optional graph model orders record choices; the source fill choices
remain deterministic. No weights changed. Local training and separate fill
holdouts remain separate denominators. The partial budget outcome, including
its wrong exact large integer, is retained. No host CPU sampling or controlled
speed comparison was performed.

`observe.go` independently recounts the original inputs/expectations and exact
native values, keeping JSON integers intact. It compares all assignments,
selection and caller/final outputs with the six prior feature observations at
`e115bb775e6a8a418dbb85eba13c16c606836f8f`. It also checks that the three native
platform checks reproduce the budget runs. `observations.tar.gz` retains all six
outputs and the three native source-fill outputs/receipt; it excludes binaries
and model weights. Extract before running:

```sh
go run ./docs/research/release-0614-20261009/observe.go \
  <this-observation-directory> examples/caller-source-fill \
  <prior-source-fill-observation-directory>
```

Copy `build.json` alongside the extracted outputs. The prior raw outputs are
also public in the workbench's `examples/caller-source-fill/observations.tar.gz`.
The test/race logs cover the release validators and CLI checks. The new platform
validator uses original feature outputs as fixtures and rejects altered hole
assignments, candidate/source identities, local and holdout counts, exact values,
replay flags and denominators. Existing native release checks remain included.
