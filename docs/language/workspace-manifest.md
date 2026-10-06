# Gooo workspace package graph

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
checks that wiring. `package execute` takes a finite case file, lowers the
entry's explicit producer chain into a typed activity graph, generates each
body in that chain, compiles the resulting Go, and runs it twice. Unrelated
activities in the workspace are left out of the execution graph. Case keys use
`<package-path>:<activity>`;
root inputs use that key, multi-input roots append `.<input-port>`, and expected
outputs use the activity key. The receipt preserves package-to-lowered-activity
identity and reports observed finite-case accuracy. This path requires 2 to 16
activities on the entry's explicitly connected producer chain and supported
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
