# Gooo workspace package graph

## Reuse a selected package program

Save `package execute --json` output, then use `package replay --receipt <saved.json>`
with the workspace manifest and either `--cases` or `--inputs`. The
[diagnostic tool example](../../examples/package-diagnostic-replay/README.md)
shows a Gooo library whose three body choices can be selected by the own compact
model and subsequently reused without the model files.

Replay rereads the manifest and its current sources, reconstructs the dependency
graph, source fills and external fill plans, and verifies the saved native
projection. It compiles and runs that program twice on the requested inputs.
The selected body stays fixed; changed input rows do not initiate another search.
Copying the same manifest-relative source tree to another location is supported.
Changed declarations, dependency identities or entry points need a new execution
record. Older external-fill records that lack `external_plan` need one new
`package execute` using their original plan; source-owned records contain their
construction contract already.

With `--json`, `replayed_from_sha256` identifies the exact supplied receipt and
`result.replay.prior_result_sha256` binds its decoded result. Historical generation
records retain their old model-call counts. This invocation's replay and runtime
model-call counts are zero; provider environment settings are not consulted.
Fresh case results report `PASS` or `PROGRESS`; input-only results report
`OBSERVED` and keep correctness unknown. These decisions describe execution and
finite expectations, not correctness across all inputs. Prior runtime observations
are retained in the input receipt, rather than reused as current measurements.

## Resolve and execute

`gooo.workspace.json` connects source files to the package graph already used
by Gooo's package runtime. Resolve a workspace with:

```sh
gooo package resolve gooo.workspace.json
gooo package resolve --json gooo.workspace.json
gooo package execute --json --cases examples/package-imports/cases.json \
  --body-plans examples/package-imports/body-plans.json \
  examples/package-imports/gooo.workspace.json
```

The cross-package binding example at `examples/package-imports` uses this
command directly. The workspace reader accepts regular `.gooo` sources and
`.gooo.fixture` files for runnable examples that are not part of the language
conformance corpus:

```sh
gooo package resolve --json examples/package-imports/gooo.workspace.json
```

Without `--json`, resolution prints each package's imports and exported
entities and activities, then lists checked imported-activity bindings. This
view helps inspect a workspace before consuming its JSON receipt. Both forms
resolve the package graph only; they do not run activity bodies.

Version 1 records an entry activity and packages. Each package has a stable
path, a Gooo package name, imports by package path, and source paths relative to
the workspace manifest:

```json
{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "boundedint", "activity": "Clamp"},
  "packages": [
    {
      "path": "boundedint",
      "name": "boundedint",
      "imports": [],
      "sources": ["main.gooo"]
    }
  ]
}
```

Source files can state their package dependencies directly:

```gooo
package app
namespace app
import "example/core"
```

An alias can name a package in an explicit activity binding:

```gooo
import core "example/core"
bind core.Normalize.result -> Main.input
```

Resolution reads the listed `.gooo` sources, builds a deterministic dependency
order, and emits a `gooo/package-runtime-result/v1` receipt with source and
semantic digests. Unknown package imports and dependency cycles fail closed.
The JSON receipt also reports the selected entry and the current zero-effect
boundary.

Each package result now includes its statically resolved `exports` list.
Entities are exported with their stable IDs; activity exports bind every input
and output entity type to that ID. An activity can use an entity in
its own package or an entity declared in one of its direct imports. Unknown or
ambiguous imported types fail closed. When an entity exists locally, the local
declaration takes precedence over imported entities with the same name.

Record names are scoped to the owning package during executable lowering.
Different packages may each declare `Result` with distinct stable IDs and field
layouts; the compiler assigns distinct execution names and records the mapping
in `program.entity_aliases`. Constructors, signatures and supported assembly
expressions use the resolved type. Names inside strings, comments, field labels
and ordinary local-variable references retain their original spelling. Aliases
of the same stable record ID share one declaration when their ordered fields
agree. The [two-Result workspace](../../examples/package-record-namespaces/README.md)
exercises this behavior through real package execution and exact integer values.
Ambiguous imported constructor names need an unambiguous declaration/import
environment, as do signature types. Scalar entities keep the current
Integer/Boolean/Text body profiles.
When source files declare imports, their union must match the manifest's
`imports` list. A mismatch fails with `PACKAGE_SOURCE_IMPORT_MISMATCH` so source
intent and the workspace graph cannot quietly drift apart. Existing workspaces
that keep imports only in the manifest remain supported until their source is
updated.

Gooo source imports identify package dependencies, while the workspace manifest
still supplies package paths, source files, and the executable entry. Activity
call expressions across packages are not yet part of the language. `package
resolve` can resolve an explicit imported-activity binding against the imported
package's exported activity and the local consumer. The output and input must
carry the same stable entity ID. Its receipt records producer and consumer
packages, activities, ports, and entity ID under the consumer package's
`bindings` list. Unknown aliases, activities, ports, and type mismatches fail
closed. Imported producers can feed local consumers. `package resolve` only
checks that wiring. `package execute` takes a finite case file or an input-only
document, lowers the entry's explicit producer chain into a typed activity graph, generates each
body in that chain, compiles the resulting Go, and runs it twice. Unrelated
activities in the workspace are left out of the execution graph. Case keys use
`<package-path>:<activity>`;
root inputs use that key, multi-input roots append `.<input-port>`, and expected
outputs use the activity key. The receipt preserves package-to-lowered-activity
identity and reports observed finite-case accuracy when expectations are supplied.
This path supports 1 to 16 activities: the entry runs independently, or its producers
join through explicit binds. It accepts supported
scalar or declared-record values. Ordinary bodies follow their declared Gooo
bodies deterministically; `--assembly-model` is an optional local model for assembling source-declared
body plans. A source-owned `source_fill` plan in an activity's Gooo
`assembling` block supplies typed body-hole candidates directly from source; an
optional `--body-plans` file remains available for activities without such a
declaration. An activity cannot receive both plan forms. Gooo scores each
complete candidate against its declared cases, then Laya may select among
those listed candidates through `GOOO_LAYA_URL`; with no provider configured,
selection stays deterministic. Gooo fills the chosen body before package code
generation and execution, and the receipt records each fill's input-source
digest, selected candidate, and finite-case score. `--tiny-model <model.json>`
loads the local compact model once and uses it for sequential fills; for a
multi-hole plan, its current mapping is limited to distinct supported root
operations in each candidate's first declared hole. This option cannot be
mixed with Laya endpoint or credential settings. Without a configured
provider, selection remains deterministic. The model call runs synchronously after the typed plan is
built and before the generated package is compiled. To connect Laya, set
`GOOO_LAYA_URL` to its `/v1/systemone` endpoint for the same command. Fill
receipts record provider decision time; native build and run observations
record wall time, CPU time, and peak resident memory when the host exposes them.

For an ordinary language tool, use `--inputs <inputs.json>` instead of `--cases`.
The input document has schema `gooo/body-composition-inputs/v1` and an `inputs`
array of named root-input maps. No expected output is supplied. Plain output is
one JSON entry result per row; `--json` includes the complete receipt with
`decision: OBSERVED` and `inputs_digest`. Two matching native executions establish
replay, while finite expectation counts remain 0/0 and correctness remains
unobserved. The existing `--cases` mode still requires nonempty expectations.
See the [standalone Gooo assembly explainer](../../examples/assembly-explainer/README.md).

Declared records are parsed with the explicit EntityFields V4 profile throughout
package resolution, IR lowering, and workspace flattening. V4 preserves optional
single string, Boolean and Integer fields across workspace binds: omission stays
absent, while explicit `""`, `false` and `0` stay present. `null` remains invalid.
The runnable [optional-record workspace](../../examples/package-optional-record-flow/README.md)
checks absent fields, explicit zero values and a large exact integer through the
full package execution path. For example,
`examples/package-record-flow` sends a `Candidate` record from an imported
`Submit` activity into an app-owned body that constructs a typed `Review` record.
Run it with:

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/package-record-flow/cases.json \
  examples/package-record-flow/gooo.workspace.json
```

Finite cases check the actual record values at both activities, then runtime
replay checks that the generated Go projection produces the same result again.
EntityFields V3 currently covers required single string, boolean, and integer
fields in the original record-flow example. The workspace runtime's V4 profile
also supports optional single fields of those types. Repeated, nested, and
arbitrary user-defined field types remain outside this support boundary.
