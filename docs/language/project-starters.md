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
