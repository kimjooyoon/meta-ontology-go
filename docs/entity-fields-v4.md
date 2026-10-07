# EntityFields V4 optional scalar fields

`gooo check`, `gooo discover`, `gooo generate` and `gooo run --record-input` use
the exact EntityFields V4 profile. V4 preserves the required single string,
Boolean and Integer fields from V3 and adds optional single fields of those same
scalar types. The Go projection uses `*string`, `*bool` and `*int64` for
optional values; `nil` means absent, while a non-nil pointer to `""`, `false`
or `0` preserves an explicitly supplied zero value.

Field order, stable IDs, spans and the V4 profile digest remain bound in the
generated source map. The generation manifest binds that map's digest to the
generated output and semantic IR. The checked-in
[V4 example](../examples/entity-fields-v4/main.gooo.fixture) shows the supported
shape. `gooo generate` can also compile explicit activity bindings into a
runtime-plan artifact under V4. For `run --record-input`, omitted optional keys
remain absent across declared `record.forward` edges; supplied `""`, `false`
and `0` remain present values. Required keys must be supplied, extra keys and
JSON `null` are rejected before any activity is applied. This route transports
data and does not execute authored record-body code.

`required × many`, `optional × many`, nested record fields and non-scalar
optional values remain unsupported. Pure record-body codegen and
`body-compose` accept optional single scalar fields as typed pointers for
copy-and-transport bodies. Missing keys remain absent across explicit binds;
`""`, `false` and `0` remain present. Learned field assembly does not yet
synthesize new optional values. `gooo package resolve` and `gooo package execute`
also accept V4 fields throughout workspace parsing, IR lowering and flattening.
The [optional package-flow example](../examples/package-optional-record-flow/README.md)
checks that absent fields and explicit zero values survive an imported activity
bind and native replay.
