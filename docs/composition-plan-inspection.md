# Read the inputs before assembling

The development `body-plan` command lists the values a caller must supply, the
activities that will run, and any bodies that need assembly first. It uses the
same structural plan as `body-compose`. Start here when an activity has several
inputs or calls an assembling helper.

From this compiler checkout, build the development binary and inspect a graph:

```sh
go build -trimpath -o .gooo ./cmd/gooo
./.gooo body-plan --source examples/body-codegen/native-input-joins.gooo.fixture --entry Label
```

The caller supplies `Left`, `Right`, `Label.input1` and `Label.input2`. `Add` gets
its two inputs from `Left` and `Right`; `Label.input0` gets the result of `Add`.
Unrelated declarations are excluded when `--entry Label` selects that graph.
Without `--entry`, the plan includes all declared graph activities.

## Make an input file

```sh
./.gooo body-plan --source examples/body-codegen/native-input-joins.gooo.fixture \
  --entry Label --inputs-template > inputs.json
```

The file uses the existing `gooo/body-composition-inputs/v1` schema:

```json
{
  "schema": "gooo/body-composition-inputs/v1",
  "inputs": [{"Left": 0, "Right": 0, "Label.input1": false, "Label.input2": ""}]
}
```

Replace the zeros, `false` and empty strings with your inputs. These are editable
placeholders. Required record fields get the same typed placeholders; optional
fields are omitted so that absence stays explicit. Add an optional field when it
is present, including when its value is zero, `false` or an empty string.

```sh
./.gooo body-compose --source examples/body-codegen/native-input-joins.gooo.fixture \
  --entry Label --inputs inputs.json --out out/observed-label
```

The output directory must be new. Input-only execution records actual values and
keeps correctness unmeasured. Supply your own expected outputs with `--cases`
on saved replay when you want to check them. See [observe and replay](native-body-composition.md).

## Use the plan in tools

`--json` emits `gooo/body-composition-inspection/v1`. Its `source_sha256` binds
the exact source bytes. `plan` retains the typed-plan digest, semantic fingerprint,
record layouts, stable activity IDs and bind edges. `caller_inputs` exposes the
exact native input keys, scalar kinds or record entity IDs. A single-input root
uses the activity name; a multiple-input root uses `Activity.inputN`.

`assemblies` lists called bodies first, in prerequisite order, followed by graph
activities requiring assembly. A body that is both called and bound appears
once. Each item records its source contract digest, profile, source case counts
and whether it depends on another assembling body. Helper arguments remain
inside the caller's body; they do not become extra caller input keys.

The export performs zero model calls, candidate tests and native executions.
It describes the source structure. Generation separately checks body types,
source-local cases and candidates. `body-context --model` separately inspects
model compatibility. A structural plan can therefore exist for a body that
later fails a type check. Tools must retain the source digest and recompute the
plan after source changes before using it as the basis for generation.

This is a development addition after v0.6.22-dev. Public 0.6.22 binaries do not
contain this command. The saved assembly feedback workbench currently supports
one record assembly with fixed consumers; this export supplies the metadata
needed to extend that path to several assemblies and additional root inputs.
