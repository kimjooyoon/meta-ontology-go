# Gooo project starters

`gooo init` creates a small project that can be checked and used immediately.
The default `app` starter is a single activity. Choose `library` to start with
two Gooo packages connected by an imported activity binding:

```sh
gooo init --template library --module example.org/team/forecast forecast
cd forecast
gooo package resolve gooo.workspace.json
gooo package execute --json --cases cases.json gooo.workspace.json
```

## Start a reusable diagnostic tool

The `diagnostic` starter creates a Gooo tool that turns counts and diagnostic
text into a suggested next operation. Its six files are embedded in the CLI:

```sh
gooo init --template diagnostic my-diagnostic
cd my-diagnostic
gooo package execute --json --cases cases.json gooo.workspace.json > execution.json
gooo package replay --receipt execution.json --inputs inputs.json gooo.workspace.json
```

`diagnostics.gooo` owns the conditions, assignments, three record-field choices
and five construction cases. `app.gooo` imports and formats the result. The
supplied runtime suite checks eight activity outputs across four input rows.
`inputs.json` provides three additional rows for actual use, without expected
answers. Editing these input rows does not require another candidate search.
The replayed program returns values such as
`partial: missing branch result [repair-and-replay]` with zero new model calls.

Choose a package prefix with `--module example.org/team/diagnostic`; this binds
the imports, workspace paths, input keys and semantic IDs. The default prefix
is `diagnostic`. You can move the complete project and its receipt together.
Existing directories are preserved; start in a new directory.

The default construction uses deterministic candidate ordering. Add
`--assembly-model /path/to/model.json` to the first `package execute` invocation
to use a compatible local three-field model. The generated README links a pinned
2,072-parameter own-model release. Each attempt retains its finite case and field
results. The suggested operation is returned as data for the caller to consume.

The [recorded starter use](../research/diagnostic-starter-20261008/summary.json)
created a custom-module project using the CLI, assembled it with deterministic
ordering and the existing own model, then moved the project before replay.
Both construction routes and saved replay matched 8/8 activity outputs. The
model route tried two candidates and deterministic ordering four; its first
proposal matched only 3/5 construction cases. Three input-only rows then returned
diagnostic messages with zero new model calls. This is one local paired task;
the raw receipts retain its exact source, model identity and measurement scope.

### Keep the installed compiler identity with the result

`gooo version --build --json` includes the executable's observed main-module
path, version and Go module checksum under `module`. Native execution and replay
retain those coordinates under `producer_module`; a declared module replacement
has a separate `replacement` entry. Missing build metadata omits the field, and
an unavailable version or checksum stays empty or absent.

This is embedded build metadata. `producer_source_sha` keeps its existing Git
meaning: a module-installed executable can carry a precise module version and
checksum while its Git revision remains `UNBOUND_LOCAL_SOURCE`. A short revision
inside a Go pseudo-version is not expanded into a full commit. Saved replay
records the current executable's module separately from the prior execution.
These fields do not change the declared choices, finite scores or acceptance
rules. The Go version running native child programs remains a separate value.

## Library construction and model choices

To use the local compact model for the source-declared assignment, pass its
model metadata file:

```sh
gooo package execute --json --cases cases.json --tiny-model path/to/model.json \
  gooo.workspace.json
```

The `--module` value becomes the workspace package prefix and source identity.
The library template declares `Normalize(Integer) -> Integer` and
`Clamp(Integer) -> Integer` in `example.org/team/forecast/core`, then imports
core and binds Clamp to `Main` in `example.org/team/forecast/app`. Normalize
feeds Clamp through a second explicit bind. Normalize's `assembling` block
declares two typed holes that form one assignment: a base expression and a
step. It also lists complete candidate assignments and finite examples.
`gooo package execute` reads that plan from the Gooo source, fills the holes
together, checks each candidate, and runs cases below, inside, and above the
clamp interval through the full chain. The examples provide a small regression
set, not a proof over all inputs.

The Gooo sources define the package APIs, import, activity binding, and the
body-fill plan itself. Gooo scores each candidate before emission. When a local
Laya endpoint is configured, the model can choose only among the complete
assignments in the source. `--tiny-model` is an explicit local model option
that loads once for the sequential fills; it cannot be combined with Laya
environment settings. For Normalize, the compact model maps a supported root
operation in the first hole to a distinct complete assignment; it does not
jointly reason over later holes. Without either model provider, Gooo follows the
deterministic declared candidate order. Gooo checks the selected body, compiles
the generated Go, runs the named package case, and reports a replayable receipt
with finite observed accuracy. Unsupported or ambiguous plan shapes produce a
diagnostic.

The workspace manifest binds source files, package identity, entry, and
dependency order. `gooo package resolve` prints the checked graph; `gooo
package execute` follows the explicit producer chain and runs its generated
native program. Package distribution through a registry is not part of this
starter. See the [workspace package graph](workspace-manifest.md) contract.

Native execution needs Go 1.27.1 for the same operating system and architecture
as the Gooo executable. For example, a `darwin/amd64` Gooo binary running under
Rosetta needs a `darwin/amd64` Go toolchain, even on an Apple Silicon Mac. If Go
is installed elsewhere or the default `go` points to another architecture, pass
the matching executable explicitly:

```sh
gooo package execute --json --cases cases.json --go /path/to/matching/go gooo.workspace.json
```
