# {{module}}

This starter is a small Gooo package graph. The core package declares
Normalize(Integer) -> Integer and Clamp(Integer) -> Integer; the app package
imports core and exposes Main. The explicit chain increments an input, clamps
the result to the closed interval 0..10, and returns it. Normalize's
`assembling` block declares a complete two-hole assignment: a base expression
and a step. The candidates and their finite examples live in Gooo source, so
the workspace executes without a separate body-plan JSON file.

Install the Gooo CLI and run these commands from this directory:

~~~sh
go install github.com/kimjooyoon/meta-ontology-go/cmd/gooo@dev
gooo package resolve gooo.workspace.json
gooo package execute --json --cases cases.json gooo.workspace.json
~~~

To use the local compact model for the source-declared assignment:

~~~sh
gooo package execute --json --cases cases.json --tiny-model path/to/model.json \
  gooo.workspace.json
~~~

`--tiny-model` loads one local model for the sequential activity fills. It cannot
be combined with `GOOO_LAYA_URL` or `GOOO_LAYA_API_KEY`. Without it, Laya is
used when configured; otherwise Gooo follows the deterministic declared
candidate order. Laya may select only one of the complete assignments in the
Gooo source. The compact model maps a supported root operation in Normalize's
first hole to a distinct assignment; it does not jointly reason over later
holes or invent fills.

The execute command follows both declared bindings, fills Normalize's two
holes, compiles the generated Go, and runs three cases through the whole chain.
The cases cover values below, inside, and above the clamp range. Gooo scores
each candidate before emission, checks the selected body, and reports a
replayable receipt with finite observed accuracy. These examples do not prove
behavior for all integer inputs. Unsupported or ambiguous compact-model plan
shapes produce a diagnostic.

The workspace manifest records package imports and the public entry. Package
resolve prints the deterministic graph receipt; package execute adds generated
native execution and a replayable receipt for the chosen body fill.
