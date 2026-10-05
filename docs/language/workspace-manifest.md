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

This connects workspace source files to package identity and dependency
ordering. It does not yet let a Gooo source file import or call another
package's declarations; the `imports` list is currently package graph input,
not a source-level name-resolution feature. The receipt resolves declarations
and entry shape; it does not claim to execute generated Go code. The library
starter includes a workspace manifest so package boundaries are visible from
the first project command.
