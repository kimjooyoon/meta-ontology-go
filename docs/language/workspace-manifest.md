# Gooo workspace package graph

`gooo.workspace.json` connects source files to the package graph already used
by Gooo's package runtime. Resolve a workspace with:

```sh
gooo package resolve gooo.workspace.json
gooo package resolve --json gooo.workspace.json
```

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

The manifest `imports` list is the current source of package dependencies;
Gooo source syntax does not yet contain an import declaration or an activity
call expression. Type resolution therefore checks activity signatures against
the package's own entities and its directly imported packages. The receipt
resolves declarations and the public type surface; it does not claim to execute
generated Go code. The library starter includes a workspace manifest so package
boundaries are visible from the first project command.
