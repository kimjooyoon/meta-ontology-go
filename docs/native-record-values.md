# Record values inside generated activity bodies

Record constructors can declare source-owned field alternatives with per-field
completion observations. See [record field assembly](record-field-assembly.md).

An entity's existing `fields` declaration can describe a value carried by pure
activity bodies. A body can read a field, construct a complete named record,
copy it to a local or return it. `body-compose` accepts required single string
and Boolean fields, passes the actual generated value through an explicit typed
`bind`, and records each field's stable ID and typed value.

Think of the record as a small labelled tray: each field has a place and an
identity. The connection passes the whole tray to the next activity. The native
program uses a Go value struct in declaration field order. JSON is used at the
case/observation boundary.

## Declare and use the fields

```gooo
entity Candidate id "records://candidate" fields {
    field title id "records://candidate/title" type string required one
    field state id "records://candidate/state" type string required one
}
activity Propose(Integer, Text) -> Candidate computes `
if input0 >= 0 {
    return Candidate{title: input1, state: "ready"}
} else {
    return Candidate{title: input1, state: "wait"}
}`
activity Echo(Candidate) -> Candidate computes "return input"
bind Propose.result -> Echo.input
```

Use named keys for declared fields. Field names are case sensitive and may be
lowercase or uppercase identifiers. Required scalar fields must be supplied;
optional single scalar fields may be omitted. The profile supports up to 16
records and 16 fields per record. Repeated fields and nested record fields
remain unsupported. Under EntityFields V2, `boolean required one` takes a
Boolean value. EntityFields V3 adds `integer required one`, represented as Go
`int64`. EntityFields V4 adds optional string, Boolean and integer fields,
represented as `*string`, `*bool` and `*int64` with `omitempty` JSON tags.
Scalar activity parameters and results continue to use Integer, Boolean and
Text.

`input.title` reads a field. A local copy such as `let copy = input` can be
replaced with a new complete record. Parameters stay read-only; assignments
target an existing whole local. Equality compares all record fields. Calls,
loops and external effects keep their separate language boundaries.

Stable entity and field IDs determine exported native names. JSON keys keep
source field names. The report's `record_types` retains source names, stable
IDs, scalar type IDs and generated names. Each route is typechecked
and compared after restoring source names; the receipt binds the record
contract and function signature. Body generation and composition use the
separately versioned EntityFields V3 profile. Existing V1 and V2 entry points
retain their string-only and string/Boolean contracts.

## Construct and execute the complete example

The [source](../examples/body-codegen/native-records.gooo.fixture) has six
activities: a selectable Integer body, a two-input record constructor, a record
and Boolean conditional, a text label, a record identity and a two-record
comparison. [Six cases](../examples/body-codegen/native-records-cases.json) include
empty text, Korean/newline text and both int64 bounds. They provide 36 named
output expectations; omitted intermediate expectations receive no score credit.

```sh
gooo body-compose \
  --source examples/body-codegen/native-records.gooo.fixture \
  --cases examples/body-codegen/native-records-cases.json \
  --out record-results
```

Provide `--model /path/to/model.json` to rank Score's three binary choices with
the existing local three-choice model. Other bodies follow their Gooo source.
A disconnected request uses deterministic search. The same source can mix these
operations without attaching record types to the scalar function. Saved
composition replay makes zero new predictions.

External record inputs and expected record outputs must be JSON objects with
exact source field names. Required fields must appear; an omitted optional
field is absent. JSON `null` is rejected for optional fields so absence is not
confused with a supplied value. String values are limited to 1,024 UTF-8 bytes
per field; Boolean values must be JSON `true` or `false`; Integer values must
be exact signed 64-bit JSON integers. JSON member order is presentation only.
Duplicate, extra or incorrectly typed fields fail before model loading.

`runtime.json` includes `actual_fields` for record results and `input_fields`
or each input port's `fields` for record inputs. Each entry retains `id`, `name`
and actual `value` in declaration order. Activity/producer IDs and named
expectation counts remain available. Wrong expected values lower the finite
score while the executed program and actual values remain visible.

These observations describe the authored cases and source profile. Registered
`run` activity implementations and domain handoff operations keep their own
runtime contracts. Record-valued learned assembly and cross-invocation
feedback remain further work.

## Typed fields across source-driven execution

`gooo run --record-input` uses EntityFields V4. Required and optional single
`string`, `boolean` and `integer` fields are decoded from JSON with exact scalar
types and checked against their source declarations before any activity
executes. An omitted optional key stays absent in each result's `fields`; an
explicit empty string, `false` or `0` stays present. Extra keys and JSON `null`
are rejected. The compiled field schema includes stable type IDs and presence
in its operation digest. V1-V3 parser and lowering entry points remain
available to callers that need earlier profiles.

The public `gooo check` and `gooo generate` commands use EntityFields V4, which
also projects optional single scalar fields to Go pointers. Pure record-body
codegen and `body-compose` can copy these values and carry them over explicit
binds. Runtime field observations retain the stable field ID and an explicit
present flag, so absence remains distinct from an empty string, `false` or `0`.
This bounded path does not synthesize optional values through learned field
assembly.

The runnable [Boolean record-binding source](../examples/language-record-binding/boolean.gooo.fixture)
and [input](../examples/language-record-binding/boolean-input.json) carry a
`Complete` Boolean alongside the existing strings. This separate example
exercises source-declared record transport while preserving the V1 syntax
corpus fixture.

The [optional record-binding source](../examples/language-record-binding/optional.gooo.fixture)
and [input](../examples/language-record-binding/optional-input.json) omit
`Note`, while preserving an explicit empty `Label`, `Complete: false` and
`Count: 0` through a typed record edge. Try it with:

```sh
gooo run --json --entry Capture --record-input \
  examples/language-record-binding/optional-input.json \
  examples/language-record-binding/optional.gooo.fixture
```

The EntityFields V3 profile added required single `integer` fields to the Go
projection and source-driven body-generation and `body-compose` paths. V4
extends ordinary pure record-body generation and `body-compose` to optional
single scalar fields for typed copy and transport, preserving absence separately
from explicit zero values.

The [optional body source](../examples/body-codegen/optional-record-transport.gooo.fixture)
and [cases](../examples/body-codegen/optional-record-transport-cases.json)
exercise absent fields, explicit `""`/`false`/`0`, partial presence and an
integer beyond JavaScript's exact range across a declared bind:

```sh
gooo body-compose \
  --source examples/body-codegen/optional-record-transport.gooo.fixture \
  --cases examples/body-codegen/optional-record-transport-cases.json
```

Absent optional values are omitted from generated JSON. `null`, wrong scalar
types, duplicate keys and undeclared fields fail during case validation before
the model provider is loaded.

```gooo
entity Boolean id "booleans://boolean"
entity Gate id "booleans://gate" fields {
    field enabled id "booleans://gate/enabled" type boolean required one
}
activity Build(Boolean) -> Gate computes "return Gate{enabled: false}" assembling {
    choice "enabled" field_value at "0" alternative "input" intent "전달한 판단값을 보존한다. Preserve the supplied decision."
    value_case "[true]" -> "{\"enabled\":true}"
    value_case "[false]" -> "{\"enabled\":false}"
    attempts "2"
}
```

Try the checked-in example:

```sh
gooo body-codegen --json --activity Build \
  examples/body-codegen/boolean-records.gooo.fixture
```

The generated record uses a Go `bool`; finite-case observations retain JSON
booleans and the field receipt carries the stable Boolean type ID. The optional
local model can rank only the declared field-value alternative. Without it, the
same bounded search uses deterministic ordering. `gooo generate` also projects
this profile to a Go field of type `bool`. In `body-compose`, external record
input, bound record delivery, actual field observations and expected outputs
all retain Boolean values as JSON booleans.

The [composition example](../examples/body-codegen/boolean-record-composition.gooo.fixture)
passes both `true` and `false` through an explicitly bound record:

```sh
gooo body-compose \
  --source examples/body-codegen/boolean-record-composition.gooo.fixture \
  --cases examples/body-codegen/boolean-record-composition-cases.json \
  --out boolean-record-results
```

The two cases expect four values across `Build` and `Relay`; a successful run
reports `finite_passed: 4`, `finite_total: 4`, and `model_calls: 0` for runtime
replay.
