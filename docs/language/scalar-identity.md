# Scalar names and value identity

Gooo's native value profile recognizes three canonical type IDs:

| Identity | Value representation | Example name |
| --- | --- | --- |
| `urn:gooo:type:integer` | signed 64-bit integer | `정수` or `Count` |
| `urn:gooo:type:boolean` | Boolean | `논리` or `Enabled` |
| `urn:gooo:type:string` | UTF-8 string | `문자열` or `Message` |

The declaration's name belongs to source code. Its recognized ID describes the
native value. Changing a name while keeping that ID lets the generated function
use the same representation. The typed composition plan retains the authored
names and entity IDs. The compiler resolves them before model selection.

```gooo
entity 정수 id "urn:gooo:type:integer"
activity Add(정수, 정수) -> 정수 computes "return input0 + input1"
```

Ordinary references, manifest entries and source hashes still describe the
edited source, so each edited version needs fresh receipts. Supplied integer
values remain exact through validation and native delivery. Boolean inputs
require JSON booleans, and Text values retain the profile's 1,024-byte limit.

Existing `Integer`, `Boolean` and `Text` declarations with other nominal IDs
retain their legacy representations. A recognized canonical ID takes precedence
over its display spelling. Other names with unknown nominal IDs require a
separately supported value layout. Explicit record blocks, including `{}`, do not
supply a scalar representation. A canonical scalar cannot also declare fields.

## Compose a small program with the own model

The public `examples/scalar-identity` program declares three Korean type names.
Its body calculates a local string using an integer condition and a Boolean.
An assembling choice can place that calculated string in a record's text field.
Four authored assembly cases guide selection; four different runtime inputs
observe the selected program, including exact large integers and an emoji.

```sh
go build -trimpath -o .gooo ./cmd/gooo
./.gooo body-compose --source examples/scalar-identity/source.gooo.fixture \
  --cases examples/scalar-identity/cases.json --entry Describe --out out/scalar-fixed
```

Add `--model /absolute/path/to/graph-model.json` to use a compatible own graph
chooser. The model orders the source-declared choices; cases check the selected
body. Omitting that option uses deterministic order. Save separate outputs for
the two modes and compare their observations.

Replay uses `out/scalar-fixed/original.gooo` with the same external cases and
`--composition out/scalar-fixed/composition.json`. It reconstructs the selected
program and runs with zero new inference. The counts describe the supplied
assembly examples and runtime inputs separately.
