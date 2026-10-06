# EntityFields V1 support view

This section is generated from the profile-bound observation and observed public routes; it is not a language-completeness score.

profile_id=gooo.entityfields.go-projection.v1
profile_version=1
meta_source=internal/meta/entityfields/entity-fields-meta.gooo
fixture=examples/entity-fields-v1/main.gooo

| Surface | State | Scope |
| --- | --- | --- |
| ordinary/default parser | DEFERRED | default parser remains deferred until explicit V1 opt-in |
| explicit V1 parser | SUPPORTED | profile-bound syntax parser on the canonical fixture |
| formatter | SUPPORTED | syntax formatting and canonical replay |
| semantic lowerer | SUPPORTED | profile-bound semantic IR lowering |
| BX Get/Put | SUPPORTED | source-preserving and semantic mutation round trips |
| Go generator | SUPPORTED | profile-bound structural Go projection |
| source map | SUPPORTED | generated structural source-map projection |
| LSP adapter | SUPPORTED | EntityFieldsSyntaxParser adapter evidence only |
| public CLI check | SUPPORTED | gooo check --semantic on the canonical fixture |
| public CLI generate | SUPPORTED | gooo generate and exact generated artifact comparison |
| generated Go | SUPPORTED | generated module compilation only, not business-behavior proof |
| LSP server/editor support | UNKNOWN | no server or editor integration claim without separate evidence |

## Optional scalar fields in the public CLI

`gooo check`, `gooo discover` and `gooo generate` use the exact EntityFields V4 profile. V4
preserves the required single string, Boolean and Integer fields from V3 and
adds optional single fields of those same scalar types. The Go projection uses
`*string`, `*bool` and `*int64` for optional values; `nil` means absent, while a
non-nil pointer to `""`, `false` or `0` preserves an explicitly supplied zero
value. Field order, stable IDs, spans and the V4 profile digest remain bound in
the generated source map; the generation manifest binds that map's digest to
the generated output and semantic IR.

The checked-in [V4 example](../examples/entity-fields-v4/main.gooo.fixture)
shows the supported shape. `required × many`, `optional × many`, nested record
fields and non-scalar optional values remain unsupported. `gooo generate` can
also compile explicit activity bindings into a runtime-plan artifact under V4.
The pure record-body codegen and record-input runtime keep their V3 contract;
optional record values cannot yet be executed on those routes.
