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

To use the local compact arithmetic model for both activity bodies, pass its
model metadata file:

```sh
gooo package execute --json --cases cases.json --body-plans body-fill-plans.json \
  --tiny-model path/to/model.json gooo.workspace.json
```

The library template declares `Normalize(Integer) -> Integer` in
`boundedint/core`, then imports and binds that activity to `Main` in
`boundedint/app`. Normalize's two typed holes form one assignment: a base
expression and a step. The plan offers complete candidate assignments and
finite examples; Gooo fills the holes together, checks each candidate, and runs
the package case from `Normalize(7)` through `Main`. The examples provide a
small regression set, not a proof over all inputs.

The Gooo sources define the package APIs, import, and activity binding. Gooo
scores each candidate before emission. When a local Laya endpoint is
configured, the model can choose only among the candidates already declared in
the plan. `--tiny-model` is an explicit local model option that loads once for
the sequential fills; it cannot be combined with Laya environment settings.
For Normalize, the compact model maps a supported root operation in the first
hole to a distinct complete assignment; it does not jointly reason over later
holes. Without either model provider, Gooo follows the deterministic declared
candidate order. Gooo checks the selected bodies, compiles the generated Go,
runs the named cases, and reports a replayable receipt with finite observed
accuracy. Unsupported or ambiguous plan shapes produce a diagnostic.

The workspace manifest binds source files, package identity, entry, and
dependency order. `gooo package resolve` prints the checked graph; `gooo
package execute` follows the explicit producer chain and runs its generated
native program. Package distribution through a registry is not part of this
starter. See the [workspace package graph](workspace-manifest.md) contract.
