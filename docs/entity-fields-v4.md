# EntityFields V4 optional scalar fields

`gooo check`, `gooo discover` and `gooo generate` use the exact EntityFields V4
profile. V4 preserves the required single string, Boolean and Integer fields
from V3 and adds optional single fields of those same scalar types. The Go
projection uses `*string`, `*bool` and `*int64` for optional values; `nil` means
absent, while a non-nil pointer to `""`, `false` or `0` preserves an explicitly
supplied zero value.

Field order, stable IDs, spans and the V4 profile digest remain bound in the
generated source map. The generation manifest binds that map's digest to the
generated output and semantic IR. The checked-in
[V4 example](../examples/entity-fields-v4/main.gooo.fixture) shows the supported
shape. `gooo generate` can also compile explicit activity bindings into a
runtime-plan artifact under V4.

`required × many`, `optional × many`, nested record fields and non-scalar
optional values remain unsupported. Pure record-body codegen and
`gooo run --record-input` retain their V3 contract, so those execution paths do
not yet accept optional record values.
