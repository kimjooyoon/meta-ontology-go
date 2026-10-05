# Gooo project starters

`gooo init` creates a small project that can be checked and used immediately.
The default `app` starter is a single activity. Choose `library` to start with
two Gooo packages connected by an imported activity binding:

```sh
gooo init --template library boundedint
cd boundedint
gooo package resolve gooo.workspace.json
gooo package execute --json --cases cases.json --body-plans body-fill-plans.json gooo.workspace.json
```

The library template declares `Normalize(Integer) -> Integer` in
`boundedint/core`, then imports and binds that activity to `Main` in
`boundedint/app`. Each activity has a typed expression hole. The plan file
lists candidate expressions and finite examples for each body. The workspace
case starts at `Normalize(7)` and expects the normalized value to reach `Main`.
The examples provide a small regression set, not a proof over all inputs.

The Gooo sources define the package APIs, import, and activity binding. Gooo
scores each candidate before emission. When a local Laya endpoint is
configured, the model can choose only among the candidates already declared in
the plan. Without Laya, selection is deterministic. Gooo checks the selected
bodies, compiles the generated Go, runs the named cases, and reports a
replayable receipt with finite observed accuracy.

The workspace manifest binds source files, package identity, entry, and
dependency order. `gooo package resolve` prints the checked graph; `gooo
package execute` follows the explicit producer chain and runs its generated
native program. Package distribution through a registry is not part of this
starter. See the [workspace package graph](workspace-manifest.md) contract.
