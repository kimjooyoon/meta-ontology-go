# Record values inside generated activity bodies

Record constructors can declare source-owned field alternatives with per-field
completion observations. See [record field assembly](record-field-assembly.md).

An entity's existing `fields` declaration can describe a value carried by pure
activity bodies. A body can read a field, construct a complete named record,
copy it to a local or return it. `body-compose` currently accepts required
single-string fields, passes the actual generated value through an explicit
typed `bind`, and records each field's stable ID and value.

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

Use every declared field exactly once, with named keys. Field names are case
sensitive and may be lowercase or uppercase identifiers. Each `string required
one` field takes a Text value. The profile supports up to 16 records and 16
fields per record. Optional/many fields and nested record fields need subsequent
entity-profile work. Scalar activity parameters and results continue to use
Integer, Boolean and Text.

`input.title` reads a field. A local copy such as `let copy = input` can be
replaced with a new complete record. Parameters stay read-only; assignments
target an existing whole local. Equality compares all record fields. Calls,
loops and external effects keep their separate language boundaries.

Stable entity and field IDs determine exported native names. JSON keys keep
source field names. The report's `record_types` retains source names, stable
IDs, nominal string type IDs and generated names. Each route is typechecked
and compared after restoring source names; the receipt binds the record
contract and function signature. Pure body entry points select the existing
V1 field profile. Other frontend entry points keep their explicit field
activation contract.

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

External record inputs and expected record outputs must be complete JSON
objects with exact source field names and explicit strings of at most 1,024
UTF-8 bytes per field. JSON member order is presentation only. Missing,
duplicate, extra, null or incorrectly typed fields retain an error before
model loading.

`runtime.json` includes `actual_fields` for record results and `input_fields`
or each input port's `fields` for record inputs. Each entry retains `id`, `name`
and actual `value` in declaration order. Activity/producer IDs and named
expectation counts remain available. Wrong expected values lower the finite
score while the executed program and actual values remain visible.

These observations describe the authored cases and source profile. Registered
`run` activity implementations and domain handoff operations keep their own
runtime contracts. Record-valued learned assembly and cross-invocation
feedback remain further work.

## Boolean fields in Go projection and body generation

The separate EntityFields V2 profile adds required single `boolean` fields to
the compiler's Go projection and source-driven body-generation path. V1 remains
unchanged. The current `body-compose` record transport still accepts its
original string-only contract.

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
this profile to a Go field of type `bool`. `body-compose` record transport is
still V1 and string-only.
